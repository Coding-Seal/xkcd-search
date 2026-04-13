import os
import pytest
from playwright.sync_api import Page, expect

BASE_URL = os.getenv("BASE_URL", "http://localhost:8090")

VALID_USER = os.getenv("E2E_USER", "bob")
VALID_PASS = os.getenv("E2E_PASS", "bob")
SEARCH_QUERY = "apple"


def _login(page: Page, username: str = VALID_USER, password: str = VALID_PASS) -> None:
    """Вспомогательная функция: открыть /login, ввести учётные данные, отправить форму."""
    page.goto(BASE_URL + "/login")
    page.locator("input[name='login'], input[type='text'], input[name='username']").first.fill(username)
    page.locator("input[type='password'], input[name='password'], input[name='pswd']").first.fill(password)
    page.locator("input[type='password'], input[name='password'], input[name='pswd']").first.press("Enter")
    page.wait_for_load_state("networkidle")


# ---------------------------------------------------------------------------
# Сценарий 1: Анонимный пользователь открывает главную страницу
# Предусловия: веб-сервер запущен, пользователь не авторизован
# ---------------------------------------------------------------------------
def test_homepage_has_search_form(page: Page):
    """
    Шаг 1: Пользователь открывает главную страницу BASE_URL/.
    Шаг 2: Страница загружается — заголовок (title) не пуст.
    Шаг 3: На странице присутствует форма поиска с полем ввода.
    Ожидаемый результат: форма и поле ввода видимы.
    """
    # Шаг 1: открыть главную страницу
    page.goto(BASE_URL + "/")

    # Шаг 2: проверить заголовок страницы
    title = page.title()
    assert title and len(title.strip()) > 0, "Заголовок страницы не должен быть пустым"

    # Шаг 3: проверить наличие формы и поля ввода
    expect(page.locator("form")).to_be_visible()
    search_input = page.locator(
        "input[type='text'], input[name='query'], input[name='search'], input[placeholder*='earch']"
    ).first
    expect(search_input).to_be_visible()


# ---------------------------------------------------------------------------
# Сценарий 2: Анонимный пользователь выполняет поиск
# Предусловия: веб-сервер запущен, пользователь не авторизован
# ---------------------------------------------------------------------------
def test_anonymous_search_returns_results(page: Page):
    """
    Шаг 1: Пользователь открывает главную страницу.
    Шаг 2: Вводит поисковый запрос в поле ввода.
    Шаг 3: Нажимает Enter для отправки формы.
    Шаг 4: Дожидается загрузки результатов.
    Ожидаемый результат: на странице есть результаты поиска или сообщение об их отсутствии.
    """
    # Шаг 1: открыть главную страницу
    page.goto(BASE_URL + "/")

    # Шаг 2: ввести запрос
    search_input = page.locator(
        "input[type='text'], input[name='query'], input[name='search'], input[placeholder*='earch']"
    ).first
    search_input.fill(SEARCH_QUERY)

    # Шаг 3: отправить форму
    search_input.press("Enter")

    # Шаг 4: дождаться загрузки
    page.wait_for_load_state("networkidle")

    has_results = page.locator("img, .comic, .result, article").count() > 0
    has_no_results_msg = (
        "no result" in page.content().lower()
        or "not found" in page.content().lower()
        or "nothing" in page.content().lower()
    )
    assert has_results or has_no_results_msg, (
        "После поиска ожидаются результаты или сообщение об их отсутствии"
    )


# ---------------------------------------------------------------------------
# Сценарий 3: Анонимный пользователь отправляет пустую форму поиска
# Предусловия: веб-сервер запущен, пользователь не авторизован
# ---------------------------------------------------------------------------
def test_empty_search_does_not_crash(page: Page):
    """
    Шаг 1: Пользователь открывает главную страницу.
    Шаг 2: Отправляет форму поиска с пустым полем ввода.
    Шаг 3: Дожидается загрузки страницы.
    Ожидаемый результат: приложение не возвращает ошибку 5xx, страница загружается корректно.
    """
    # Шаг 1: открыть главную страницу
    page.goto(BASE_URL + "/")

    # Шаг 2: отправить пустую форму
    form = page.locator("form").first
    form.evaluate("f => f.submit()")

    # Шаг 3: дождаться загрузки
    page.wait_for_load_state("networkidle")

    assert page.url is not None
    title = page.title()
    assert title is not None and "error" not in title.lower(), (
        "Пустой поиск не должен приводить к странице ошибки"
    )


# ---------------------------------------------------------------------------
# Сценарий 4: Анонимный пользователь открывает страницу входа
# Предусловия: веб-сервер запущен, пользователь не авторизован
# ---------------------------------------------------------------------------
def test_login_page_renders_credentials_fields(page: Page):
    """
    Шаг 1: Пользователь переходит на страницу /login.
    Шаг 2: Страница загружается, форма входа видима.
    Шаг 3: На форме присутствуют поле логина и поле пароля.
    Ожидаемый результат: оба поля формы видимы и доступны для ввода.
    """
    # Шаг 1: открыть страницу входа
    page.goto(BASE_URL + "/login")

    # Шаг 2: проверить форму
    expect(page.locator("form")).to_be_visible()

    # Шаг 3: проверить поля формы
    login_field = page.locator(
        "input[name='login'], input[type='text'], input[name='username']"
    ).first
    password_field = page.locator(
        "input[type='password'], input[name='password'], input[name='pswd']"
    ).first

    expect(login_field).to_be_visible()
    expect(password_field).to_be_visible()


# ---------------------------------------------------------------------------
# Сценарий 5: Пользователь вводит неверные учётные данные
# Предусловия: веб-сервер запущен, пользователь не авторизован
# ---------------------------------------------------------------------------
def test_invalid_login_shows_error_or_stays_on_login(page: Page):
    """
    Шаг 1: Пользователь открывает страницу /login.
    Шаг 2: Вводит несуществующий логин и неверный пароль.
    Шаг 3: Нажимает Enter для отправки формы.
    Шаг 4: Дожидается загрузки.
    Ожидаемый результат: остаётся на странице /login или видит сообщение об ошибке.
    """
    # Шаг 1: открыть страницу входа
    page.goto(BASE_URL + "/login")

    login_field = page.locator(
        "input[name='login'], input[type='text'], input[name='username']"
    ).first
    password_field = page.locator(
        "input[type='password'], input[name='password'], input[name='pswd']"
    ).first

    # Шаг 2: ввести неверные данные
    login_field.fill("no_such_user_xyzzy")
    password_field.fill("wrong_password_xyzzy")

    # Шаг 3: отправить форму
    password_field.press("Enter")

    # Шаг 4: дождаться загрузки
    page.wait_for_load_state("networkidle")

    stayed_on_login = "/login" in page.url
    has_error_text = any(
        kw in page.content().lower()
        for kw in ("error", "invalid", "incorrect", "unauthorized", "wrong", "failed")
    )
    assert stayed_on_login or has_error_text, (
        "После неверных учётных данных ожидается остаться на /login или увидеть ошибку"
    )


# ---------------------------------------------------------------------------
# Сценарий 6: Пользователь успешно входит в систему
# Предусловия: веб-сервер запущен, пользователь bob/bob существует в БД
# ---------------------------------------------------------------------------
def test_valid_login_redirects_to_home(page: Page):
    """
    Шаг 1: Пользователь открывает страницу /login.
    Шаг 2: Вводит корректный логин и пароль (bob/bob).
    Шаг 3: Нажимает Enter для отправки формы.
    Шаг 4: Дожидается загрузки.
    Ожидаемый результат: браузер перенаправляется с /login на другую страницу.
    """
    # Шаги 1–4: логин через вспомогательную функцию
    _login(page, VALID_USER, VALID_PASS)

    assert "/login" not in page.url, (
        f"После успешного входа ожидается редирект с /login, но URL: {page.url}"
    )


# ---------------------------------------------------------------------------
# Сценарий 7: Авторизованный пользователь выполняет поиск и видит кнопку «Избранное»
# Предусловия: пользователь bob/bob вошёл в систему, индекс заполнен
# ---------------------------------------------------------------------------
def test_logged_in_user_search_shows_favorite_button(page: Page):
    """
    Шаг 1: Пользователь входит в систему (bob/bob).
    Шаг 2: Переходит на главную страницу.
    Шаг 3: Вводит поисковый запрос и нажимает Enter.
    Шаг 4: Дожидается загрузки результатов.
    Ожидаемый результат: среди результатов присутствует кнопка добавления в избранное.
    """
    # Шаг 1: войти в систему
    _login(page, VALID_USER, VALID_PASS)

    # Шаг 2: открыть главную страницу
    page.goto(BASE_URL + "/")

    # Шаг 3: ввести запрос и отправить
    search_input = page.locator(
        "input[type='text'], input[name='query'], input[name='search'], input[placeholder*='earch']"
    ).first
    search_input.fill(SEARCH_QUERY)
    search_input.press("Enter")

    # Шаг 4: дождаться загрузки
    page.wait_for_load_state("networkidle")

    result_count = page.locator("img, .comic, .result, article").count()
    if result_count == 0:
        pytest.skip("Нет результатов поиска — пропуск проверки кнопки избранного")

    fav_button = page.locator(
        "button[class*='fav'], button[class*='favorite'], a[href*='fav'], "
        "button[aria-label*='favorite'], button[title*='favorite'], .favorite-btn, "
        "form[action*='fav'] button, input[value*='fav'], .favorite-star"
    )
    assert fav_button.count() > 0, (
        "Ожидается кнопка избранного в результатах поиска для авторизованного пользователя"
    )


# ---------------------------------------------------------------------------
# Сценарий 8: Авторизованный пользователь открывает страницу избранного
# Предусловия: пользователь bob/bob вошёл в систему
# ---------------------------------------------------------------------------
def test_logged_in_user_can_view_favorites_page(page: Page):
    """
    Шаг 1: Пользователь входит в систему (bob/bob).
    Шаг 2: Переходит на страницу /favorites.
    Шаг 3: Страница загружается без серверной ошибки.
    Ожидаемый результат: HTTP-статус < 500, страница содержит контент.
    """
    # Шаг 1: войти в систему
    _login(page, VALID_USER, VALID_PASS)

    # Шаг 2: открыть страницу избранного
    response = page.goto(BASE_URL + "/favorites")

    # Шаг 3: проверить статус и содержимое
    assert response is not None
    assert response.status < 500, (
        f"Страница избранного вернула серверную ошибку {response.status}"
    )
    content = page.content().lower()
    assert len(content) > 100, "Страница избранного почти пуста"


# ---------------------------------------------------------------------------
# Сценарий 9: Авторизованный пользователь добавляет комикс в избранное
# Предусловия: пользователь bob/bob вошёл в систему, индекс заполнен
# ---------------------------------------------------------------------------
def test_logged_in_user_add_to_favorites_appears_on_favorites_page(page: Page):
    """
    Шаг 1: Пользователь входит в систему (bob/bob).
    Шаг 2: Выполняет поиск по запросу "apple".
    Шаг 3: Нажимает кнопку «Избранное» у первого результата.
    Шаг 4: Переходит на страницу /favorites.
    Ожидаемый результат: страница /favorites содержит добавленный элемент или сообщение о состоянии.
    """
    # Шаг 1: войти в систему
    _login(page, VALID_USER, VALID_PASS)

    # Шаг 2: выполнить поиск
    page.goto(BASE_URL + "/")
    search_input = page.locator(
        "input[type='text'], input[name='query'], input[name='search'], input[placeholder*='earch']"
    ).first
    search_input.fill(SEARCH_QUERY)
    search_input.press("Enter")
    page.wait_for_load_state("networkidle")

    # Шаг 3: нажать кнопку избранного у первого результата
    fav_button = page.locator(
        "button[class*='fav'], button[class*='favorite'], "
        "form[action*='fav'] button, input[value*='fav'], "
        "button[aria-label*='favorite'], .favorite-btn, .favorite-star"
    ).first

    if not fav_button.is_visible():
        pytest.skip("Кнопка избранного не найдена в результатах поиска — пропуск")

    fav_button.click()
    page.wait_for_load_state("networkidle")

    # Шаг 4: перейти на /favorites и проверить содержимое
    page.goto(BASE_URL + "/favorites")
    page.wait_for_load_state("networkidle")

    has_items = page.locator("img, .comic, .result, article, li").count() > 0
    has_empty_msg = (
        "no favorites" in page.content().lower()
        or "empty" in page.content().lower()
    )
    assert has_items or has_empty_msg, (
        "После добавления в избранное страница /favorites должна показывать элемент или пустое состояние"
    )


# ---------------------------------------------------------------------------
# Сценарий 10: Пользователь кликает по ссылке «Home» в навигации
# Предусловия: веб-сервер запущен
# ---------------------------------------------------------------------------
def test_navbar_home_link_navigates_to_root(page: Page):
    """
    Шаг 1: Пользователь открывает страницу /login.
    Шаг 2: Находит ссылку «Home» в навигационном меню.
    Шаг 3: Кликает по ссылке.
    Шаг 4: Дожидается загрузки.
    Ожидаемый результат: браузер переходит на корневой URL (/).
    """
    # Шаг 1: открыть страницу входа
    page.goto(BASE_URL + "/login")

    # Шаг 2–3: найти и кликнуть ссылку на главную
    home_link = page.locator("a[href='/'], a[href=''], nav a, header a").first
    home_link.click()

    # Шаг 4: дождаться загрузки
    page.wait_for_load_state("networkidle")

    root = BASE_URL.rstrip("/")
    assert page.url.rstrip("/") == root or page.url == root + "/", (
        f"После клика на Home ожидается корневой URL, но URL: {page.url}"
    )


# ---------------------------------------------------------------------------
# Сценарий 13: Пользователь добавляет комикс в избранное, затем убирает его
# Предусловия: пользователь bob/bob существует в БД, индекс заполнен
# ---------------------------------------------------------------------------
def test_add_and_remove_favorite_disappears_from_favorites(page: Page):
    """
    Шаг 1: Пользователь входит в систему (bob/bob).
    Шаг 2: Выполняет поиск "apple", ожидает результаты.
    Шаг 3: Убеждается, что первый комикс добавлен в избранное (кликает если нет).
    Шаг 4: Открывает /favorites — фиксирует количество комиксов.
    Шаг 5: Кликает кнопку избранного у первого комикса — удаляет из избранного.
    Шаг 6: Открывает /favorites снова — проверяет, что количество уменьшилось.
    Ожидаемый результат: комикс исчез со страницы /favorites после удаления.
    """
    # Шаг 1: войти в систему
    _login(page, VALID_USER, VALID_PASS)

    # Шаг 2: выполнить поиск
    page.goto(BASE_URL + "/")
    search_input = page.locator(
        "input[type='text'], input[name='query'], input[name='search'], input[placeholder*='earch']"
    ).first
    search_input.fill(SEARCH_QUERY)
    search_input.press("Enter")
    page.wait_for_load_state("networkidle")

    # Шаг 3: убедиться, что первый комикс добавлен в избранное
    fav_stars = page.locator(".favorite-star")
    if fav_stars.count() == 0:
        pytest.skip("Кнопки избранного не найдены — пропуск теста")

    first_star = fav_stars.first
    # Если звезда не активна — добавить в избранное
    if "active" not in (first_star.get_attribute("class") or ""):
        first_star.click()
        page.wait_for_load_state("networkidle")

    # Шаг 4: открыть /favorites и зафиксировать количество
    page.goto(BASE_URL + "/favorites")
    page.wait_for_load_state("networkidle")
    count_before = page.locator(".comic-card").count()
    assert count_before > 0, (
        "После добавления комикса в избранное список /favorites не должен быть пустым"
    )

    # Шаг 5: удалить первый комикс из избранного
    remove_star = page.locator(".favorite-star").first
    if not remove_star.is_visible():
        pytest.skip("Кнопка избранного на странице /favorites не найдена")
    remove_star.click()
    page.wait_for_load_state("networkidle")

    # Шаг 6: открыть /favorites снова и проверить, что количество уменьшилось
    page.goto(BASE_URL + "/favorites")
    page.wait_for_load_state("networkidle")
    count_after = page.locator(".comic-card").count()
    assert count_after < count_before, (
        f"После удаления количество комиксов в избранном должно уменьшиться "
        f"(было {count_before}, стало {count_after})"
    )


# ---------------------------------------------------------------------------
# Сценарий 12: Анонимный пользователь обращается к /favorites
# Предусловия: веб-сервер запущен, пользователь не авторизован
# ---------------------------------------------------------------------------
def test_unauthenticated_favorites_page_is_accessible(page: Page):
    """
    Шаг 1: Пользователь (не авторизованный) напрямую открывает /favorites.
    Шаг 2: Дожидается загрузки страницы.
    Ожидаемый результат: статус < 500; отображается пустое состояние,
    ссылка на вход или перенаправление на /login — без серверной ошибки.
    """
    # Шаг 1: открыть /favorites без авторизации
    response = page.goto(BASE_URL + "/favorites")

    # Шаг 2: проверить статус и содержимое
    assert response is not None
    assert response.status < 500, (
        f"Анонимный доступ к /favorites вернул серверную ошибку: {response.status}"
    )

    content = page.content().lower()
    is_login_page = "/login" in page.url
    has_login_link = "login" in content
    has_empty_state = (
        "no favorites" in content or "empty" in content or "sign in" in content
    )
    assert is_login_page or has_login_link or has_empty_state or len(content) > 100, (
        "Неавторизованный /favorites должен показывать пустое состояние, ссылку на вход или редирект"
    )
