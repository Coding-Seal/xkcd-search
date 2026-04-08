import json
import argparse
import requests

parser = argparse.ArgumentParser("port")
parser.add_argument("port", type=int)
args = parser.parse_args()

print("Start pics test...")
host = f"http://localhost:{args.port}"

# Step 1: login to get JWT token
print("Logging in...")
credentials = {"login": "bob", "password": "bob"}
resp = requests.post(host + "/api/login", data=json.dumps(credentials))
if resp.status_code != 200:
    print("Failed to login")
    print(resp)
    exit(1)
jwt = resp.headers.get("Authorization")

# Step 2: search for comics
query = "apple,doctor"
expected_url = "https://imgs.xkcd.com/comics/an_apple_a_day.png"
print("Fetching pics...")
resp = requests.get(f"{host}/api/pics?search={query}", headers={"Authorization": jwt})
if resp.status_code != 200:
    print("Failed to fetch pics")
    print(resp)
    exit(1)
pics = json.loads(resp.text)
if not any(p.get("img") == expected_url for p in pics):
    print("Expected comic not found in pics response")
    print(resp.text)
    exit(1)

# Step 3: get specific comic by ID from search results
comic_id = pics[0]["id"]
print(f"Fetching comic by id={comic_id}...")
resp = requests.get(f"{host}/api/comic/{comic_id}", headers={"Authorization": jwt})
if resp.status_code != 200:
    print("Failed to fetch comic by id")
    print(resp)
    exit(1)
comic = json.loads(resp.text)
if comic.get("id") != comic_id:
    print("Comic id mismatch")
    print(resp.text)
    exit(1)

print("pics test PASS")
