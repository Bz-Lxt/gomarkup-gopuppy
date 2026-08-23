#!/usr/bin/env python3
"""API smoke in Mock mode. Expected cost ¥0."""
import json
import os
import urllib.error
import urllib.request

API = os.environ.get("API_URL", "http://127.0.0.1:27152").rstrip("/")


def req(method, path, body=None, token=None, status=200):
    data = None if body is None else json.dumps(body).encode()
    r = urllib.request.Request(API + path, data=data, method=method)
    r.add_header("Content-Type", "application/json")
    if token:
        r.add_header("Authorization", "Bearer " + token)
    try:
        with urllib.request.urlopen(r, timeout=15) as resp:
            raw = resp.read()
            code = resp.status
    except urllib.error.HTTPError as e:
        raw = e.read()
        code = e.code
    assert code == status, f"{method} {path} -> {code} {raw[:200]!r} want {status}"
    return json.loads(raw.decode() or "null")


def main():
    hz = urllib.request.urlopen(API + "/healthz", timeout=8)
    assert hz.status == 200
    login = req("POST", "/api/v1/auth/login", {"email": "dad@gopuppy.test", "password": "Puppy123!"})
    token = login["data"]["tokens"]["access_token"]
    families = req("GET", "/api/v1/families", token=token)["data"]
    assert families, "no family"
    fid = families[0]["id"]
    pets = req("GET", f"/api/v1/families/{fid}/pets", token=token)["data"]
    assert len(pets) >= 2
    cream = next(p for p in pets if p["name"] == "奶油")
    assert cream["age"]["total_days"] > 0
    fin = req("GET", f"/api/v1/pets/{cream['id']}/finance", token=token)["data"]
    assert fin["weight_series"] and fin["expense_series"]
    ev = req("GET", f"/api/v1/pets/{cream['id']}/events", token=token)["data"]
    assert any(e["category"] == "SURGERY" for e in ev)
    # cross-family 404: random uuid
    req("GET", "/api/v1/pets/77777777-7777-7777-7777-777777777777", token=token, status=404)
    # checkin idempotent
    pid = cream["id"]
    req("POST", f"/api/v1/pets/{pid}/checkins", {"type": "FEED", "slot": "NOON", "done": True}, token)
    req("POST", f"/api/v1/pets/{pid}/checkins", {"type": "FEED", "slot": "NOON", "done": True}, token)
    today = req("GET", f"/api/v1/pets/{pid}/checkins/today", token=token)["data"] or []
    noon = [c for c in today if c.get("slot") == "NOON" and c.get("type") == "FEED" and not c.get("revoked_at")]
    assert len(noon) <= 1
    print("SMOKE_OK")


if __name__ == "__main__":
    main()
