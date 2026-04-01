#!/usr/bin/env python3
import json
import os
import sys
import urllib.parse
import urllib.request
import urllib.error
from typing import Any

BASE_URL = os.getenv("YOUGILE_BASE_URL", "https://yougile.com").rstrip("/")
LOGIN = os.getenv("YOUGILE_LOGIN", "").strip()
PASSWORD = os.getenv("YOUGILE_PASSWORD", "")
COMPANY_ID_ENV = os.getenv("YOUGILE_COMPANY_ID", "").strip()
TASK_LIMIT = int(os.getenv("YOUGILE_TASK_LIMIT", "200"))


def fail(message: str, code: int = 1) -> None:
    print(f"ERROR: {message}", file=sys.stderr)
    raise SystemExit(code)


def as_items(payload: Any) -> list[dict[str, Any]]:
    if isinstance(payload, dict):
        content = payload.get("content")
        if isinstance(content, list):
            return [x for x in content if isinstance(x, dict)]
    if isinstance(payload, list):
        return [x for x in payload if isinstance(x, dict)]
    return []


def request_json(method: str, path: str, body: dict[str, Any] | None = None, token: str | None = None) -> Any:
    url = f"{BASE_URL}{path}"
    data = None
    headers = {"Accept": "application/json"}

    if body is not None:
        data = json.dumps(body).encode("utf-8")
        headers["Content-Type"] = "application/json"

    if token:
        headers["Authorization"] = f"Bearer {token}"

    req = urllib.request.Request(url=url, data=data, method=method.upper(), headers=headers)

    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            raw = resp.read().decode("utf-8", errors="replace").strip()
            if not raw:
                return {}
            return json.loads(raw)
    except urllib.error.HTTPError as e:
        raw = e.read().decode("utf-8", errors="replace").strip()
        fail(f"HTTP {e.code} {method} {path}: {raw}")
    except urllib.error.URLError as e:
        fail(f"Network error {method} {path}: {e}")


def discover_companies(login: str, password: str) -> list[dict[str, Any]]:
    payload = request_json("POST", "/api-v2/auth/companies?limit=1000", {
        "login": login,
        "password": password,
    })
    return as_items(payload)


def create_key(login: str, password: str, company_id: str) -> str:
    payload = request_json("POST", "/api-v2/auth/keys", {
        "login": login,
        "password": password,
        "companyId": company_id,
    })
    key = str(payload.get("key") or payload.get("apiKey") or "").strip()
    if not key:
        fail("create key response has no key/apiKey field")
    return key


def list_users(api_key: str) -> list[dict[str, Any]]:
    payload = request_json("GET", "/api-v2/users?limit=1000", token=api_key)
    return as_items(payload)


def list_tasks(api_key: str, assigned_to: str | None = None) -> list[dict[str, Any]]:
    params = {
        "limit": str(TASK_LIMIT),
        "includeDeleted": "false",
    }
    if assigned_to and assigned_to.strip():
        params["assignedTo"] = assigned_to.strip()

    suffix = urllib.parse.urlencode(params)
    payload = request_json("GET", f"/api-v2/task-list?{suffix}", token=api_key)
    return as_items(payload)


def normalize_assigned(task: dict[str, Any]) -> list[str]:
    assigned = task.get("assigned")
    if not isinstance(assigned, list):
        return []
    result: list[str] = []
    for item in assigned:
        if isinstance(item, str) and item.strip():
            result.append(item.strip())
    return result


def resolve_my_user_id(users: list[dict[str, Any]], login: str) -> str:
    login_l = login.lower().strip()
    for user in users:
        email = str(user.get("email") or "").strip().lower()
        if email and email == login_l:
            uid = str(user.get("id") or "").strip()
            if uid:
                return uid

    # fallback: try exact login field if API has it
    for user in users:
        user_login = str(user.get("login") or "").strip().lower()
        if user_login and user_login == login_l:
            uid = str(user.get("id") or "").strip()
            if uid:
                return uid

    fail("cannot resolve my user id from /api-v2/users by email/login")
    return ""


def task_id(task: dict[str, Any]) -> str:
    return str(task.get("id") or "").strip()


def main() -> None:
    if not LOGIN or not PASSWORD:
        fail("YOUGILE_LOGIN and YOUGILE_PASSWORD env vars are required")

    print("[1/6] discover companies...")
    companies = discover_companies(LOGIN, PASSWORD)
    if not companies:
        fail("no companies returned for credentials")

    company_id = COMPANY_ID_ENV
    if not company_id:
        preferred = next((c for c in companies if bool(c.get("isAdmin"))), companies[0])
        company_id = str(preferred.get("id") or "").strip()
    if not company_id:
        fail("cannot resolve company id")

    print(f"    companies={len(companies)} selected_company={company_id}")

    print("[2/6] create api key...")
    api_key = create_key(LOGIN, PASSWORD, company_id)
    print(f"    api_key_last4=...{api_key[-4:]}")

    print("[3/6] list users...")
    users = list_users(api_key)
    print(f"    users={len(users)}")

    print("[4/6] resolve my user id...")
    my_user_id = resolve_my_user_id(users, LOGIN)
    print(f"    my_user_id={my_user_id}")

    print("[5/6] list tasks with server filter assignedTo=my_user_id...")
    tasks_server = list_tasks(api_key, assigned_to=my_user_id)
    print(f"    tasks_server={len(tasks_server)}")

    print("[6/6] list all tasks and client-filter by assigned includes my_user_id...")
    tasks_all = list_tasks(api_key)
    tasks_client = [t for t in tasks_all if my_user_id in normalize_assigned(t)]
    print(f"    tasks_all={len(tasks_all)} tasks_client={len(tasks_client)}")

    server_ids = {task_id(t) for t in tasks_server if task_id(t)}
    client_ids = {task_id(t) for t in tasks_client if task_id(t)}

    only_server = sorted(server_ids - client_ids)
    only_client = sorted(client_ids - server_ids)

    print("\nResult:")
    print(f"  overlap={len(server_ids & client_ids)}")
    print(f"  only_server={len(only_server)}")
    print(f"  only_client={len(only_client)}")

    if only_server:
        print("  sample_only_server:", only_server[:5])
    if only_client:
        print("  sample_only_client:", only_client[:5])

    if only_server or only_client:
        print("\nWARNING: server-side assignee filter and client-side filter differ")
    else:
        print("\nOK: server-side assignee filter matches client-side executor filter")


if __name__ == "__main__":
    main()
