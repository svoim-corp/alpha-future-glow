"""Проверка естественных уточнений, ссылок и планера на настоящем локальном ИИ.

Создаёт демо-план и архивирует его только после подтверждения. Базу не сбрасывает.
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
    parser.add_argument("--output", default="support-capabilities-results.json")
    parser.add_argument("--skip-credit", action="store_true")
    parser.add_argument("--read-only", action="store_true")
    args = parser.parse_args()
    client = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))
    report = []

    def request(path, body=None, key=None):
        headers = {"Content-Type": "application/json"}
        if key:
            headers["Idempotency-Key"] = key
        raw = None if body is None else json.dumps(body, ensure_ascii=False).encode()
        with client.open(urllib.request.Request(args.base_url + "/api/v1" + path, data=raw, headers=headers), timeout=20) as response:
            return json.load(response)

    request("/auth/login", {"login": "smirnov", "password": "demo123"})
    conversation = request("/support/conversations", {})
    path = "/support/conversations/" + conversation["id"]

    def ask(text, kind=None):
        started = time.monotonic()
        request(path + "/messages", {"text": text}, str(uuid.uuid4()))
        while time.monotonic() - started < 150:
            latest = request(path + "/messages")[-1]
            if latest["role"] == "assistant" and latest["status"] != "PENDING":
                report.append({"question": text, "seconds": round(time.monotonic()-started, 2), "message": latest})
                with open(args.output, "w", encoding="utf-8") as output:
                    json.dump(report, output, ensure_ascii=False, indent=2)
                assert latest["status"] == "COMPLETED", latest
                if kind:
                    assert any(b["kind"] == kind for b in latest["blocks"]), latest
                print(json.dumps(report[-1], ensure_ascii=False), flush=True)
                return latest
            time.sleep(1)
        raise TimeoutError(text)

    def confirm(message):
        action = next(b for b in message["blocks"] if b["kind"] == "action")
        action_path = path + "/actions/" + action["actionId"] + "/confirm"
        result = request(action_path, {}, action["actionId"])
        replay = request(action_path, {}, "repeat-" + action["actionId"])
        assert result["resultId"] == replay["resultId"]
        return result

    if args.read_only:
        identity = ask("Расскажи обо мне и моей компании", "company")
        company = next(b["company"] for b in identity["blocks"] if b["kind"] == "company")
        assert company.get("viewerName") and company.get("viewerPosition"), identity
        credits = ask("Покажи мои кредиты")
        assert not any(b["kind"] == "payment" for b in credits["blocks"])
        loans = request("/loans")
        if loans:
            assert any(b["kind"] == "loan" for b in credits["blocks"]), credits
        ask("Покажи ближайшие платежи")
        ask("Покажи реквизиты моей компании", "company")
        ask("Покажи мои обращения")
        print("Проверка чтения пройдена", flush=True)
        return

    if not args.skip_credit:
        first = ask("Дай мне кредит, хочу на 150к рублей и на 12 месяца. Хочу пару телефонов взять")
        assert "150000" in first["text"] and "12" in first["text"], first
        ask("Телефоны для сотрудников, для работы бизнеса", "product")
        ask("Рассчитай платёж по экспресс-кредиту с этими параметрами", "calculation")
        draft = ask("Подготовь черновик этого экспресс-кредита", "action")
        params = next(b["action"]["draft"] for b in draft["blocks"] if b.get("action", {}).get("kind") == "draft")
        assert params["amount"] == 150000 and params["term"] == 12, params
    navigation = ask("Открой планер", "navigation")
    assert navigation["blocks"][0]["navigation"]["href"] == "/planner"
    follow = ask("Ссылку на него дай", "navigation")
    assert follow["blocks"][0]["navigation"]["href"] == "/planner"
    title = "Проверка ассистента " + uuid.uuid4().hex[:6]
    plan = ask(f"Создай план «{title}»: купить печь для пекарни, бюджет 500 тысяч рублей, 10 ноября 2026 года", "action")
    before = request("/plans")["items"]
    assert not any(p["title"] == title for p in before), "Сохранено без подтверждения"
    result = confirm(plan)
    saved = next(p for p in request("/plans")["items"] if p["id"] == result["resultId"])
    assert saved["budgetRub"] == 500000 and saved["startDate"] == "2026-11-10", saved
    ask("Покажи мои планы", "plan")
    updated = ask(f"В плане {result['resultId']} измени бюджет на 600 тысяч рублей", "action")
    confirm(updated)
    saved = next(p for p in request("/plans")["items"] if p["id"] == result["resultId"])
    assert saved["budgetRub"] == 600000 and saved["startDate"] == "2026-11-10", saved
    archive = ask(f"Архивируй план {result['resultId']}", "action")
    confirm(archive)
    assert any(p["id"] == result["resultId"] for p in request("/plans?state=ARCHIVED")["items"])
    ask("Покажи реквизиты моей компании", "company")
    ask("Покажи мои обращения", "ticket")
    credits = ask("Покажи мои кредиты")
    assert not any(b["kind"] == "payment" for b in credits["blocks"]), "Список кредитов подменён платежами"
    print("Проверка расширенных возможностей пройдена", flush=True)


if __name__ == "__main__":
    main()
