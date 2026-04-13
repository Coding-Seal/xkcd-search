import argparse
import json
import requests

parser = argparse.ArgumentParser("port")
parser.add_argument("port", type=int)
args = parser.parse_args()
host = f"http://localhost:{args.port}"

# ---------------------------------------------------------------------------
# Сценарий 1: Администратор получает JWT и запускает обновление индекса
# Предусловия: бинарник запущен; пользователь bob/bob с ролью admin существует
# ---------------------------------------------------------------------------
print("=== Сценарий 1: Обновление индекса с авторизацией ===")

print("Шаг 1: POST /api/login — пользователь входит в систему...")
credentials = {"login": "bob", "password": "bob"}
resp = requests.post(host + "/api/login", data=json.dumps(credentials))
if resp.status_code != 200:
    print(f"FAIL: ожидался 200, получен {resp.status_code}")
    print(resp.text)
    exit(1)
jwt = resp.headers.get("Authorization")
print(f"Шаг 1 УСПЕХ: получен JWT токен")

print("Шаг 2: извлечь токен из заголовка Authorization...")
if not jwt:
    print("FAIL: заголовок Authorization отсутствует в ответе")
    exit(1)
print(f"Шаг 2 УСПЕХ: токен получен")

print("Шаг 3: POST /api/update с токеном — запрос на обновление индекса...")
resp = requests.post(host + "/api/update", headers={"Authorization": jwt})
if resp.status_code != 200:
    print(f"FAIL: ожидался 200, получен {resp.status_code}")
    print(resp.text)
    exit(1)
print(f"Шаг 3 УСПЕХ: индекс обновлён (HTTP {resp.status_code})")
print("Сценарий 1 PASS\n")

# ---------------------------------------------------------------------------
# Сценарий 2: Попытка обновления индекса без токена авторизации
# Предусловия: бинарник запущен; токен не получался
# ---------------------------------------------------------------------------
print("=== Сценарий 2: Обновление индекса без авторизации ===")

print("Шаг 1: POST /api/update без заголовка Authorization...")
resp = requests.post(host + "/api/update")
if resp.status_code != 401:
    print(f"FAIL: ожидался 401 Unauthorized, получен {resp.status_code}")
    print(resp.text)
    exit(1)
print(f"Шаг 1 УСПЕХ: получен HTTP {resp.status_code} Unauthorized — доступ закрыт")
print("Сценарий 2 PASS")
