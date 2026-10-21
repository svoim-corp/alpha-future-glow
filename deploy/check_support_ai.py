"""Сквозная проверка локального ассистента на синтетической демо-учётке.

python deploy/check_support_ai.py --base-url http://localhost:3000 --output support-ai-results.json
Создаёт один демонстрационный черновик и одно обращение. Без сторонних библиотек.
"""

import argparse
import http.cookiejar
import json
import time
import urllib.request
import uuid


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--base-url", default="http://localhost:3000")
    parser.add_argument("--output", default="support-ai-results.json")
    args = parser.parse_args()
    client = urllib.request.build_opener(
        urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar())
    )
    report = []

    def request(path, body=None, key=None):
        data = None if body is None else json.dumps(body, ensure_ascii=False).encode()
        headers = {"Content-Type": "application/json"}
        if key:
            headers["Idempotency-Key"] = key
        req = urllib.request.Request(args.base_url + "/api/v1" + path, data=data, headers=headers)
        with client.open(req, timeout=20) as response:
            return json.load(response)

    request("/auth/login", {"login": "smirnov", "password": "demo123"})
    conversation = request("/support/conversations", {})
    path = "/support/conversations/" + conversation["id"]

    def ask(text, expected_kind=None):
        started = time.monotonic()
        request(path + "/messages", {"text": text}, str(uuid.uuid4()))
        while time.monotonic() - started < 180:
            messages = request(path + "/messages")
            latest = messages[-1]
            if latest["role"] == "assistant" and latest["status"] != "PENDING":
                duration = round(time.monotonic() - started, 2)
                report.append({"question": text, "seconds": duration, "message": latest})
                with open(args.output, "w", encoding="utf-8") as output:
                    json.dump(report, output, ensure_ascii=False, indent=2)
                assert latest["status"] == "COMPLETED", latest
                if expected_kind:
                    assert any(b["kind"] == expected_kind for b in latest["blocks"]), latest
                print(json.dumps({"question": text, "seconds": duration,
                                  "blocks": [b["kind"] for b in latest["blocks"]]}, ensure_ascii=False), flush=True)
                return latest
            time.sleep(1)
        raise TimeoutError("Ассистент не завершил ответ за 180 секунд")

    def confirm(message):
        proposal = next(b for b in message["blocks"] if b["kind"] == "action")
        action_path = path + "/actions/" + proposal["actionId"] + "/confirm"
        result = request(action_path, {}, proposal["actionId"])
        replay = request(action_path, {}, "replay-" + proposal["actionId"])
        assert result["resultId"] == replay["resultId"], "Повтор создал другой объект"
        return result

    first = ask("Чем отличается овердрафт от обычного кредита?")
    assert first["sources"], "Нет справочных материалов"
    ask("Подбери кредит на покупку оборудования за 2 миллиона рублей на 24 месяца.", "product")
    ask("Рассчитай платёж по инвестиционному кредиту: 2 миллиона рублей, 24 месяца.", "calculation")
    proposal = ask("Подготовь черновик инвестиционного кредита на оборудование: 2 миллиона рублей на 24 месяца.", "action")
    result = confirm(proposal)
    applications = request("/applications?size=100")
    assert any(a["id"] == result["resultId"] for a in applications["items"])
    ask("Покажи мои заявки.", "application")
    ask("Покажи ближайшие платежи по моим кредитам.")
    ticket = ask("Подготовь обращение: в карточке компании указан старый фактический адрес склада. Нужно проверить адрес.", "action")
    result = confirm(ticket)
    assert any(t["id"] == result["resultId"] for t in request("/support/tickets"))
    print("Сквозная проверка пройдена", flush=True)


if __name__ == "__main__":
    main()
