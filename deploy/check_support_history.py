"""Проверка аккаунта, названия диалога и удаления собственного тестового черновика.

Работает с демо smirnov. Удаляет только черновик, созданный этим запуском.
"""
import http.cookiejar
import json
import time
import urllib.request
import uuid


def main():
    client = urllib.request.build_opener(
        urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar())
    )

    def request(path, body=None, method=None):
        raw = None if body is None else json.dumps(body, ensure_ascii=False).encode()
        req = urllib.request.Request(
            "http://localhost:3000/api/v1" + path, data=raw, method=method,
            headers={"Content-Type": "application/json", "Idempotency-Key": str(uuid.uuid4())},
        )
        with client.open(req, timeout=20) as response:
            data = response.read()
            return json.loads(data) if data else None

    request("/auth/login", {"login": "smirnov", "password": "demo123"})
    conversation = request("/support/conversations", {})
    path = "/support/conversations/" + conversation["id"]

    def ask(text):
        request(path + "/messages", {"text": text})
        deadline = time.monotonic() + 150
        while time.monotonic() < deadline:
            message = request(path + "/messages")[-1]
            if message["role"] == "assistant" and message["status"] != "PENDING":
                assert message["status"] == "COMPLETED", message
                return message
            time.sleep(1)
        raise TimeoutError(text)

    question = "Вот у меня как компания называется и кто я?"
    answer = ask(question)
    company = next(b["company"] for b in answer["blocks"] if b["kind"] == "company")
    assert company["viewerName"] and company["viewerPosition"]
    assert company["viewerName"] in answer["text"] and company["name"] in answer["text"]
    listed = next(c for c in request("/support/conversations") if c["id"] == conversation["id"])
    assert listed["title"] == question

    draft = request("/applications", {"productCode": "WORKING", "amount": 1500000, "term": 24})
    deleted = False
    try:
        answer = ask("Покажи мои заявки")
        card = next(b["application"] for b in answer["blocks"]
                    if b.get("application", {}).get("id") == draft["id"])
        assert card["canDelete"] is True
        request("/applications/" + draft["id"], method="DELETE")
        deleted = True
        history = request(path + "/messages")
        assert not any(b.get("application", {}).get("id") == draft["id"]
                       for m in history for b in m["blocks"])
        assert any(b["title"] == "Черновик удалён или заявка закрыта"
                   for m in history for b in m["blocks"])
    finally:
        if not deleted:
            request("/applications/" + draft["id"], method="DELETE")
    print("Проверены данные аккаунта, название диалога, удаление тестового черновика и обновление истории.")


if __name__ == "__main__":
    main()
