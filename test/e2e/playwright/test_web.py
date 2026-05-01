import os
import pytest
from playwright.sync_api import Page, expect

BASE_URL = os.getenv("BASE_URL", "http://localhost:8090")

VALID_USER = os.getenv("E2E_USER", "bob")
VALID_PASS = os.getenv("E2E_PASS", "bob")


def _login(page: Page, username: str = VALID_USER, password: str = VALID_PASS) -> None:
    """Helper: open /login, fill credentials, submit form, wait for redirect."""
    page.goto(BASE_URL + "/login")
    page.locator("input[name='login'], input[type='text']").first.fill(username)
    page.locator("input[type='password'], input[name='pswd']").first.fill(password)
    page.locator("input[type='password'], input[name='pswd']").first.press("Enter")
    page.wait_for_load_state("networkidle")


def _search(page: Page, query: str) -> None:
    """Helper: fill search input with query and submit."""
    search_input = page.locator("input[name='query'], input[type='text']").first
    search_input.fill(query)
    search_input.press("Enter")
    page.wait_for_load_state("networkidle")


def _click_star_and_wait(page: Page, star_locator) -> None:
    """Helper: click a favorite star and wait for the fetch request to complete."""
    with page.expect_response(lambda r: "/favorite" in r.url):
        star_locator.click()


# ---------------------------------------------------------------------------
# Сценарий 4: Вход с неверным паролем → остаётся на /login
# ---------------------------------------------------------------------------
def test_login_wrong_password_stays_on_login(page: Page):
    """
    Шаг 1: Открыть /login
    Шаг 2: Ввести логин "bob", пароль "wrongpassword"
    Шаг 3: Нажать Enter
    Шаг 4: Дождаться загрузки
    Ожидаемый результат: URL содержит /login; пользователь не перенаправлен
    Постусловие: cookie Authorization не установлена
    """
    page.goto(BASE_URL + "/login")
    page.locator("input[name='login'], input[type='text']").first.fill("bob")
    page.locator("input[type='password'], input[name='pswd']").first.fill("wrongpassword")
    page.locator("input[type='password'], input[name='pswd']").first.press("Enter")
    page.wait_for_load_state("networkidle")

    assert "/login" in page.url, (
        f"После неверного пароля ожидается остаться на /login, но URL: {page.url}"
    )

    cookies = {c["name"]: c for c in page.context.cookies()}
    assert "Authorization" not in cookies, (
        "Cookie Authorization не должна быть установлена после неверного пароля"
    )


# ---------------------------------------------------------------------------
# Сценарий 5: Неверный пароль → верный пароль → поиск комикса
# ---------------------------------------------------------------------------
def test_login_wrong_then_correct_password_and_search(page: Page):
    """
    Шаг 1: Открыть /login
    Шаг 2: Ввести bob/wrongpassword → Enter
    Шаг 3: Убедиться что остались на /login
    Шаг 4: Ввести bob/bob → Enter
    Шаг 5: Дождаться редиректа на главную
    Шаг 6: Ввести "apple" → Enter
    Шаг 7: Дождаться результатов
    Ожидаемый результат: после второй попытки URL не содержит /login; на главной отображаются комиксы
    Постусловие: пользователь авторизован; cookie Authorization установлена
    """
    page.goto(BASE_URL + "/login")
    page.locator("input[name='login'], input[type='text']").first.fill("bob")
    page.locator("input[type='password'], input[name='pswd']").first.fill("wrongpassword")
    page.locator("input[type='password'], input[name='pswd']").first.press("Enter")
    page.wait_for_load_state("networkidle")

    assert "/login" in page.url, "После первой неверной попытки ожидается остаться на /login"

    page.locator("input[name='login'], input[type='text']").first.fill("bob")
    page.locator("input[type='password'], input[name='pswd']").first.fill(VALID_PASS)
    page.locator("input[type='password'], input[name='pswd']").first.press("Enter")
    page.wait_for_load_state("networkidle")

    assert "/login" not in page.url, (
        f"После верного пароля ожидается редирект, но URL: {page.url}"
    )

    _search(page, "apple")

    comics = page.locator(".comic-card").count()
    assert comics > 0, "После входа поиск 'apple' должен вернуть комиксы"


# ---------------------------------------------------------------------------
# Сценарий 6: Анонимный пользователь ищет комикс — кнопок избранного нет
# ---------------------------------------------------------------------------
def test_anonymous_search_no_favorite_buttons(page: Page):
    """
    Шаг 1: Открыть /
    Шаг 2: Убедиться что присутствует форма с текстовым полем
    Шаг 3: Ввести "apple" → Enter
    Шаг 4: Дождаться загрузки результатов
    Ожидаемый результат: страница содержит элементы комиксов; кнопок .favorite-star нет
    Постусловие: состояние системы не изменилось
    """
    page.goto(BASE_URL + "/")
    expect(page.locator("form")).to_be_visible()

    _search(page, "apple")

    comics = page.locator(".comic-card").count()
    assert comics > 0, "Ожидаются результаты поиска"

    fav_stars = page.locator(".favorite-star")
    assert fav_stars.count() == 0, (
        "Анонимный пользователь не должен видеть кнопки .favorite-star"
    )


# ---------------------------------------------------------------------------
# Сценарий 7: Несуществующий запрос, затем реальный — получены результаты
# ---------------------------------------------------------------------------
def test_nonexistent_query_then_real_returns_results(page: Page):
    """
    Шаг 1: Открыть /
    Шаг 2: Ввести "xyzzy_no_such_comic_ever" → Enter
    Шаг 3: Убедиться что список пуст или нет результатов; нет 5xx ошибки
    Шаг 4: Ввести "apple" → Enter
    Шаг 5: Дождаться загрузки результатов
    Ожидаемый результат: второй поиск возвращает комиксы; оба запроса обработаны корректно
    Постусловие: состояние системы не изменилось
    """
    page.goto(BASE_URL + "/")
    _search(page, "xyzzyqqqwwwzzz")

    content = page.content().lower()
    assert "500" not in content and "internal server error" not in content, (
        "Несуществующий запрос не должен вызывать 5xx ошибку"
    )
    assert page.locator(".comic-card").count() == 0, (
        "Несуществующий запрос должен вернуть пустой список"
    )

    _search(page, "apple")

    comics = page.locator(".comic-card").count()
    assert comics > 0, "После реального запроса 'apple' ожидаются комиксы"


# ---------------------------------------------------------------------------
# Сценарий 8: Авторизованный пользователь видит кнопки избранного, анонимный — нет
# ---------------------------------------------------------------------------
def test_auth_user_sees_favorites_anon_does_not(page: Page):
    """
    Шаг 1: Открыть /, ввести "apple" → Enter
    Шаг 2: Убедиться что кнопок избранного нет (пользователь анонимный)
    Шаг 3: Открыть /login, ввести bob/bob → Enter, дождаться редиректа
    Шаг 4: Перейти на /, ввести "apple" → Enter
    Шаг 5: Дождаться результатов
    Ожидаемый результат: после входа у каждого комикса видна кнопка .favorite-star
    Постусловие: пользователь авторизован
    """
    page.goto(BASE_URL + "/")
    _search(page, "apple")

    assert page.locator(".favorite-star").count() == 0, (
        "Анонимный пользователь не должен видеть кнопки .favorite-star"
    )

    _login(page, VALID_USER, VALID_PASS)
    page.goto(BASE_URL + "/")
    _search(page, "apple")

    fav_stars = page.locator(".favorite-star")
    assert fav_stars.count() > 0, (
        "Авторизованный пользователь должен видеть кнопки .favorite-star в результатах поиска"
    )


# ---------------------------------------------------------------------------
# Сценарий 9: Войти, убедиться в пустом /favorites, найти комикс, добавить → на /favorites
# ---------------------------------------------------------------------------
def test_login_empty_favorites_add_comic_appears(page: Page):
    """
    Шаг 1: Открыть /login, ввести bob/bob → Enter
    Шаг 2: Открыть /favorites — убедиться что список пуст
    Шаг 3: Перейти на /, ввести "apple" → Enter
    Шаг 4: Запомнить title первого комикса
    Шаг 5: Нажать кнопку избранного у первого комикса
    Шаг 6: Открыть /favorites
    Ожидаемый результат: на /favorites отображается именно тот комикс из шага 4
    Постусловие: cookie Favorites содержит ID добавленного комикса
    """
    _login(page, VALID_USER, VALID_PASS)

    page.goto(BASE_URL + "/favorites")
    page.wait_for_load_state("networkidle")
    assert page.locator(".comic-card").count() == 0, (
        "Список избранного должен быть пустым в начале теста"
    )

    page.goto(BASE_URL + "/")
    _search(page, "apple")

    first_card = page.locator(".comic-card").first
    title = first_card.locator("h3").inner_text()

    first_star = page.locator(".favorite-star").first
    assert first_star.is_visible(), "Кнопка .favorite-star должна быть видна"
    _click_star_and_wait(page, first_star)

    page.goto(BASE_URL + "/favorites")
    page.wait_for_load_state("networkidle")

    assert title in page.content(), (
        f"Ожидается комикс '{title}' на странице /favorites"
    )


# ---------------------------------------------------------------------------
# Сценарий 10: Добавить два комикса, удалить один — остался только нужный
# ---------------------------------------------------------------------------
def test_add_two_remove_first_only_second_remains(page: Page):
    """
    Шаг 1: Открыть /login, ввести bob/bob → Enter
    Шаг 2: Перейти на /, ввести "apple" → Enter
    Шаг 3: Запомнить title первого и второго комикса
    Шаг 4: Нажать кнопку избранного у первого комикса
    Шаг 5: Нажать кнопку избранного у второго комикса
    Шаг 6: Открыть /favorites — убедиться что оба присутствуют
    Шаг 7: Нажать кнопку избранного у первого комикса (удаление)
    Шаг 8: Открыть /favorites
    Ожидаемый результат: на /favorites только второй комикс; первый отсутствует
    Постусловие: cookie Favorites содержит только ID второго комикса
    """
    _login(page, VALID_USER, VALID_PASS)
    page.goto(BASE_URL + "/")
    _search(page, "apple")

    cards = page.locator(".comic-card")
    assert cards.count() >= 2, "Нужно минимум 2 комикса в результатах для этого теста"

    title1 = cards.nth(0).locator("h3").inner_text()
    title2 = cards.nth(1).locator("h3").inner_text()

    _click_star_and_wait(page, page.locator(".favorite-star").nth(0))
    _click_star_and_wait(page, page.locator(".favorite-star").nth(1))

    page.goto(BASE_URL + "/favorites")
    page.wait_for_load_state("networkidle")
    content = page.content()
    assert title1 in content, f"Первый комикс '{title1}' должен быть в избранном"
    assert title2 in content, f"Второй комикс '{title2}' должен быть в избранном"

    _click_star_and_wait(page, page.locator(".favorite-star").first)

    page.goto(BASE_URL + "/favorites")
    page.wait_for_load_state("networkidle")
    content_after = page.content()
    assert title2 in content_after, f"Второй комикс '{title2}' должен остаться в избранном"
    assert title1 not in content_after, f"Первый комикс '{title1}' должен быть удалён из избранного"


# ---------------------------------------------------------------------------
# Сценарий 11: Добавить комикс в избранное, удалить — список снова пустой
# ---------------------------------------------------------------------------
def test_add_and_remove_favorite_list_becomes_empty(page: Page):
    """
    Шаг 1: Открыть /login, ввести bob/bob → Enter
    Шаг 2: Перейти на /, ввести "apple" → Enter
    Шаг 3: Нажать кнопку избранного у первого комикса (добавление)
    Шаг 4: Открыть /favorites — убедиться что комикс присутствует
    Шаг 5: Нажать кнопку избранного у того же комикса (удаление)
    Шаг 6: Открыть /favorites
    Ожидаемый результат: список избранного пуст
    Постусловие: cookie Favorites содержит пустой массив
    """
    _login(page, VALID_USER, VALID_PASS)
    page.goto(BASE_URL + "/")
    _search(page, "apple")

    first_star = page.locator(".favorite-star").first
    assert first_star.is_visible(), "Кнопка .favorite-star должна быть видна"
    _click_star_and_wait(page, first_star)

    page.goto(BASE_URL + "/favorites")
    page.wait_for_load_state("networkidle")
    assert page.locator(".comic-card").count() > 0, "Комикс должен появиться в избранном"

    remove_star = page.locator(".favorite-star").first
    assert remove_star.is_visible(), "Кнопка удаления из избранного должна быть видна"
    _click_star_and_wait(page, remove_star)

    page.goto(BASE_URL + "/favorites")
    page.wait_for_load_state("networkidle")
    count = page.locator(".comic-card").count()
    assert count == 0, f"Список избранного должен быть пустым после удаления, но найдено {count} элементов"


# ---------------------------------------------------------------------------
# Сценарий 12: Добавить комикс, выполнить новый поиск — избранное сохранилось
# ---------------------------------------------------------------------------
def test_add_favorite_new_search_favorite_persists(page: Page):
    """
    Шаг 1: Открыть /login, ввести bob/bob → Enter
    Шаг 2: Перейти на /, ввести "apple" → Enter
    Шаг 3: Нажать кнопку избранного у первого комикса (добавление)
    Шаг 4: Ввести "linux" → Enter
    Шаг 5: Дождаться новых результатов
    Шаг 6: Открыть /favorites
    Ожидаемый результат: комикс из шага 3 всё ещё присутствует; новый поиск не сбросил избранное
    Постусловие: cookie Favorites сохранена между поисками
    """
    _login(page, VALID_USER, VALID_PASS)
    page.goto(BASE_URL + "/")
    _search(page, "apple")

    first_card = page.locator(".comic-card").first
    title = first_card.locator("h3").inner_text()

    first_star = page.locator(".favorite-star").first
    assert first_star.is_visible(), "Кнопка .favorite-star должна быть видна"
    _click_star_and_wait(page, first_star)

    _search(page, "linux")

    page.goto(BASE_URL + "/favorites")
    page.wait_for_load_state("networkidle")
    assert title in page.content(), (
        f"Комикс '{title}' должен сохраниться в избранном после нового поиска"
    )


# ---------------------------------------------------------------------------
# Сценарий 13: Анонимный пользователь открывает /favorites — нет чужих данных
# ---------------------------------------------------------------------------
def test_anonymous_favorites_page_no_foreign_data(page: Page):
    """
    Шаг 1: Открыть /favorites напрямую без авторизации
    Шаг 2: Дождаться загрузки страницы
    Ожидаемый результат: статус < 500; список пуст или редирект на /login; чужих данных нет
    Постусловие: состояние системы не изменилось
    """
    response = page.goto(BASE_URL + "/favorites")
    page.wait_for_load_state("networkidle")

    assert response is not None
    assert response.status < 500, (
        f"Анонимный доступ к /favorites вернул ошибку сервера: {response.status}"
    )

    is_login_redirect = "/login" in page.url
    is_empty_list = page.locator(".comic-card").count() == 0
    assert is_login_redirect or is_empty_list, (
        "Анонимный /favorites должен показывать пустой список или перенаправить на /login"
    )


# ---------------------------------------------------------------------------
# Сценарий 14: Войти, добавить комикс, перейти на главную через navbar → избранное сохранилось
# ---------------------------------------------------------------------------
def test_navbar_home_navigation_favorite_persists(page: Page):
    """
    Шаг 1: Открыть /login, ввести bob/bob → Enter
    Шаг 2: Ввести "apple" → Enter; добавить первый комикс в избранное
    Шаг 3: Открыть /favorites — убедиться что комикс есть
    Шаг 4: Кликнуть ссылку Home в navbar
    Шаг 5: Дождаться загрузки главной страницы
    Шаг 6: Открыть /favorites снова
    Ожидаемый результат: после навигации через navbar комикс всё ещё в избранном
    Постусловие: URL равен BASE_URL/; cookie Favorites не сброшена
    """
    _login(page, VALID_USER, VALID_PASS)
    _search(page, "apple")

    first_card = page.locator(".comic-card").first
    title = first_card.locator("h3").inner_text()

    first_star = page.locator(".favorite-star").first
    assert first_star.is_visible(), "Кнопка .favorite-star должна быть видна"
    _click_star_and_wait(page, first_star)

    page.goto(BASE_URL + "/favorites")
    page.wait_for_load_state("networkidle")
    assert title in page.content(), "Комикс должен быть в избранном до навигации"

    page.locator("nav a[href='/']").click()
    page.wait_for_load_state("networkidle")

    root = BASE_URL.rstrip("/")
    assert page.url.rstrip("/") == root or page.url == root + "/", (
        f"После клика Home ожидается корневой URL, но URL: {page.url}"
    )

    page.goto(BASE_URL + "/favorites")
    page.wait_for_load_state("networkidle")
    assert title in page.content(), (
        f"Комикс '{title}' должен сохраниться в избранном после навигации через navbar"
    )


# ---------------------------------------------------------------------------
# Сценарий 15: Поисковый запрос отражается в URL после submit, следующий поиск обновляет его
# ---------------------------------------------------------------------------
def test_search_query_reflects_in_url_or_input(page: Page):
    """
    Шаг 1: Открыть /
    Шаг 2: Ввести "apple" → Enter
    Шаг 3: Проверить что "apple" присутствует в URL или в поле ввода
    Шаг 4: Ввести "linux" → Enter
    Шаг 5: Проверить что URL или поле ввода обновились на "linux"
    Ожидаемый результат: каждый поиск отражается в состоянии страницы
    Постусловие: состояние системы не изменилось
    """
    page.goto(BASE_URL + "/")
    _search(page, "apple")

    url_has_apple = "apple" in page.url
    input_value = page.locator("input[name='query']").first.input_value()
    input_has_apple = input_value == "apple"
    assert url_has_apple or input_has_apple, (
        f"После поиска 'apple' ожидается 'apple' в URL или поле ввода. URL: {page.url}, input: '{input_value}'"
    )

    _search(page, "linux")

    url_has_linux = "linux" in page.url
    input_value = page.locator("input[name='query']").first.input_value()
    input_has_linux = input_value == "linux"
    assert url_has_linux or input_has_linux, (
        f"После поиска 'linux' ожидается 'linux' в URL или поле ввода. URL: {page.url}, input: '{input_value}'"
    )
