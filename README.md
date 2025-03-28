# xkcd-search
![video](assets/example.GIF)
## Пользователи
Есть три пользователя:
* login: admin password: admin
* login: alice password: alice
* login: bob password: bob
## cURL
Получить токен
```
curl -X POST -v --data '{"login":"bob", "password":"bob"}' localhost:8080/api/login
```
Использовать токен
```
curl -v -H "Authorization: <token>" -X POST http://localhost:8080/api/update

```
## API
Все эндпоинты для работы с REST API имеют префикс /api

## Test Coverage
![Cover.svg](assets/cover.svg)

## Arch

### System Level
![System](assets/System.png)

### Container Level
![Container](assets/Container.png)
