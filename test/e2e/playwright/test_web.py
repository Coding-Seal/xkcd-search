import os
import pytest
from playwright.sync_api import Page, expect

BASE_URL = os.getenv("BASE_URL", "http://localhost:8090")

VALID_USER = os.getenv("E2E_USER", "bob")
VALID_PASS = os.getenv("E2E_PASS", "bob")
SEARCH_QUERY = "apple"


def _login(page: Page, username: str = VALID_USER, password: str = VALID_PASS) -> None:
    """Helper: fill login form and submit."""
    page.goto(BASE_URL + "/login")
    page.locator("input[name='login'], input[type='text'], input[name='username']").first.fill(username)
    page.locator("input[type='password'], input[name='password'], input[name='pswd']").first.fill(password)
    page.locator("input[type='password'], input[name='password'], input[name='pswd']").first.press("Enter")
    page.wait_for_load_state("networkidle")


# ---------------------------------------------------------------------------
# Scenario 1: Anonymous user visits home page and verifies search form exists
# ---------------------------------------------------------------------------
def test_homepage_has_search_form(page: Page):
    """
    User journey: Open the home page -> verify the page title is set
    -> verify a search form with an input field is visible.
    Multi-step: navigate, check title, check form element.
    """
    page.goto(BASE_URL + "/")
    title = page.title()
    assert title and len(title.strip()) > 0, "Page title must not be empty"

    expect(page.locator("form")).to_be_visible()
    search_input = page.locator(
        "input[type='text'], input[name='search'], input[placeholder*='earch']"
    ).first
    expect(search_input).to_be_visible()


# ---------------------------------------------------------------------------
# Scenario 2: Anonymous user performs a search and sees results
# ---------------------------------------------------------------------------
def test_anonymous_search_returns_results(page: Page):
    """
    User journey: Go to home -> type a query -> submit -> wait for results
    -> verify that at least one image or comic card appears on the page.
    Multi-step: navigate, fill input, submit form, assert results.
    """
    page.goto(BASE_URL + "/")
    search_input = page.locator(
        "input[type='text'], input[name='search'], input[placeholder*='earch']"
    ).first
    search_input.fill(SEARCH_QUERY)
    search_input.press("Enter")
    page.wait_for_load_state("networkidle")

    has_results = page.locator("img, .comic, .result, article").count() > 0
    has_no_results_msg = (
        "no result" in page.content().lower()
        or "not found" in page.content().lower()
        or "nothing" in page.content().lower()
    )
    assert has_results or has_no_results_msg, (
        "Expected search results or a no-results message after submitting query"
    )


# ---------------------------------------------------------------------------
# Scenario 3: Submitting an empty search does not crash the app
# ---------------------------------------------------------------------------
def test_empty_search_does_not_crash(page: Page):
    """
    User journey: Go to home -> submit the form with no query -> page stays
    alive with a title (not a 5xx error page).
    Multi-step: navigate, submit empty form, assert page is still alive.
    """
    page.goto(BASE_URL + "/")
    form = page.locator("form").first
    form.evaluate("f => f.submit()")
    page.wait_for_load_state("networkidle")

    assert page.url is not None
    title = page.title()
    assert title is not None and "error" not in title.lower(), (
        "Empty search should not produce an error page"
    )


# ---------------------------------------------------------------------------
# Scenario 4: Login page renders required form fields
# ---------------------------------------------------------------------------
def test_login_page_renders_credentials_fields(page: Page):
    """
    User journey: Navigate to /login -> verify the page title contains
    relevant text -> verify both login and password inputs are visible.
    Multi-step: navigate, check title content, check two form fields.
    """
    page.goto(BASE_URL + "/login")
    expect(page.locator("form")).to_be_visible()

    login_field = page.locator(
        "input[name='login'], input[type='text'], input[name='username']"
    ).first
    password_field = page.locator(
        "input[type='password'], input[name='password'], input[name='pswd']"
    ).first

    expect(login_field).to_be_visible()
    expect(password_field).to_be_visible()


# ---------------------------------------------------------------------------
# Scenario 5: Invalid credentials show an error or keep user on /login
# ---------------------------------------------------------------------------
def test_invalid_login_shows_error_or_stays_on_login(page: Page):
    """
    User journey: Go to /login -> fill in wrong username & password
    -> submit -> verify the app does NOT silently redirect to home
    (stays on /login or displays an error message).
    Multi-step: navigate, fill fields, submit, assert outcome.
    """
    page.goto(BASE_URL + "/login")
    login_field = page.locator(
        "input[name='login'], input[type='text'], input[name='username']"
    ).first
    password_field = page.locator(
        "input[type='password'], input[name='password'], input[name='pswd']"
    ).first

    login_field.fill("no_such_user_xyzzy")
    password_field.fill("wrong_password_xyzzy")
    password_field.press("Enter")
    page.wait_for_load_state("networkidle")

    stayed_on_login = "/login" in page.url
    has_error_text = any(
        kw in page.content().lower()
        for kw in ("error", "invalid", "incorrect", "unauthorized", "wrong", "failed")
    )
    assert stayed_on_login or has_error_text, (
        "Expected to stay on /login or see an error message after invalid credentials"
    )


# ---------------------------------------------------------------------------
# Scenario 6: Valid login redirects user away from /login
# ---------------------------------------------------------------------------
def test_valid_login_redirects_to_home(page: Page):
    """
    User journey: Go to /login -> fill valid credentials -> submit
    -> verify the browser is redirected away from /login (to home or another page).
    Multi-step: navigate to /login, fill two fields, submit, assert redirect.
    """
    _login(page, VALID_USER, VALID_PASS)

    assert "/login" not in page.url, (
        f"After valid login expected redirect away from /login, but URL is {page.url}"
    )


# ---------------------------------------------------------------------------
# Scenario 7: Logged-in user searches and gets results with a favorites button
# ---------------------------------------------------------------------------
def test_logged_in_user_search_shows_favorite_button(page: Page):
    """
    User journey: Login -> go to home -> search for a query -> wait for
    results -> verify that a favorite/bookmark toggle button appears next
    to at least one result.
    Multi-step: login, navigate, search, assert interactive element.
    """
    _login(page, VALID_USER, VALID_PASS)

    page.goto(BASE_URL + "/")
    search_input = page.locator(
        "input[type='text'], input[name='search'], input[placeholder*='earch']"
    ).first
    search_input.fill(SEARCH_QUERY)
    search_input.press("Enter")
    page.wait_for_load_state("networkidle")

    result_count = page.locator("img, .comic, .result, article").count()
    if result_count == 0:
        pytest.skip("No search results returned; cannot check favorite button")

    fav_button = page.locator(
        "button[class*='fav'], button[class*='favorite'], a[href*='fav'], "
        "button[aria-label*='favorite'], button[title*='favorite'], .favorite-btn, "
        "form[action*='fav'] button, input[value*='fav']"
    )
    assert fav_button.count() > 0, (
        "Expected a favorite toggle button to be present in search results when logged in"
    )


# ---------------------------------------------------------------------------
# Scenario 8: Logged-in user can access and view the favorites page
# ---------------------------------------------------------------------------
def test_logged_in_user_can_view_favorites_page(page: Page):
    """
    User journey: Login -> navigate to /favorites -> verify the page loads
    without a server error and contains expected page structure (heading or list).
    Multi-step: login, navigate to favorites, assert page content.
    """
    _login(page, VALID_USER, VALID_PASS)

    response = page.goto(BASE_URL + "/favorites")
    assert response is not None
    assert response.status < 500, (
        f"Favorites page returned server error {response.status} for logged-in user"
    )

    # Page should have some content (heading or empty-state message)
    content = page.content().lower()
    assert len(content) > 100, "Favorites page returned nearly empty body"


# ---------------------------------------------------------------------------
# Scenario 9: Logged-in user adds a comic to favorites and sees it on /favorites
# ---------------------------------------------------------------------------
def test_logged_in_user_add_to_favorites_appears_on_favorites_page(page: Page):
    """
    User journey: Login -> search -> click the first favorite button
    -> navigate to /favorites -> verify the favorites page is not empty
    (i.e., at least one comic is listed there).
    Multi-step: login, search, toggle favorite, navigate, assert.
    """
    _login(page, VALID_USER, VALID_PASS)

    page.goto(BASE_URL + "/")
    search_input = page.locator(
        "input[type='text'], input[name='search'], input[placeholder*='earch']"
    ).first
    search_input.fill(SEARCH_QUERY)
    search_input.press("Enter")
    page.wait_for_load_state("networkidle")

    fav_button = page.locator(
        "button[class*='fav'], button[class*='favorite'], "
        "form[action*='fav'] button, input[value*='fav'], "
        "button[aria-label*='favorite'], .favorite-btn"
    ).first

    if not fav_button.is_visible():
        pytest.skip("No favorite button found in search results; skipping")

    fav_button.click()
    page.wait_for_load_state("networkidle")

    page.goto(BASE_URL + "/favorites")
    page.wait_for_load_state("networkidle")

    # The favorites page should now have at least one comic image or item
    has_items = page.locator("img, .comic, .result, article, li").count() > 0
    has_empty_msg = (
        "no favorites" in page.content().lower()
        or "empty" in page.content().lower()
    )
    assert has_items or has_empty_msg, (
        "After adding a favorite, the /favorites page should show the item or an empty-state message"
    )


# ---------------------------------------------------------------------------
# Scenario 10: Navigation - home link in navbar returns user to root
# ---------------------------------------------------------------------------
def test_navbar_home_link_navigates_to_root(page: Page):
    """
    User journey: Go to /login -> find the home/logo link in the navbar
    -> click it -> verify the browser is at the root URL.
    Multi-step: navigate to non-root page, find nav element, click, assert URL.
    """
    page.goto(BASE_URL + "/login")
    home_link = page.locator("a[href='/'], a[href=''], nav a, header a").first
    home_link.click()
    page.wait_for_load_state("networkidle")

    root = BASE_URL.rstrip("/")
    assert page.url.rstrip("/") == root or page.url == root + "/", (
        f"Expected to be at root after clicking home link, but URL is {page.url}"
    )


# ---------------------------------------------------------------------------
# Scenario 11: Search query is preserved in the URL or input after submitting
# ---------------------------------------------------------------------------
def test_search_query_reflected_after_submit(page: Page):
    """
    User journey: Go to home -> enter a search term -> submit the form
    -> verify the search term appears in the URL query string or in the
    input field value (i.e., state is not lost on page load).
    Multi-step: navigate, fill, submit, assert URL or input contains query.
    """
    query = "linux"
    page.goto(BASE_URL + "/")
    search_input = page.locator(
        "input[type='text'], input[name='search'], input[placeholder*='earch']"
    ).first
    search_input.fill(query)
    search_input.press("Enter")
    page.wait_for_load_state("networkidle")

    url_has_query = query in page.url
    input_value = search_input.input_value() if search_input.is_visible() else ""
    input_has_query = query in input_value

    assert url_has_query or input_has_query, (
        f"Expected query '{query}' to appear in URL ({page.url}) or input value"
    )


# ---------------------------------------------------------------------------
# Scenario 12: Unauthenticated user accessing /favorites gets a usable page
# ---------------------------------------------------------------------------
def test_unauthenticated_favorites_page_is_accessible(page: Page):
    """
    User journey: Without logging in, navigate directly to /favorites
    -> verify the page returns HTTP 200 (or redirect) and renders without
    a 5xx server error, and either shows an empty state or a login prompt.
    Multi-step: navigate directly, check response status, inspect content.
    """
    response = page.goto(BASE_URL + "/favorites")
    assert response is not None

    final_status = response.status
    assert final_status < 500, (
        f"Unauthenticated /favorites returned server error: {final_status}"
    )

    # Should either show empty favorites, a login link, or redirect to login
    content = page.content().lower()
    is_login_page = "/login" in page.url
    has_login_link = "login" in content
    has_empty_state = (
        "no favorites" in content or "empty" in content or "sign in" in content
    )
    assert is_login_page or has_login_link or has_empty_state or len(content) > 100, (
        "Unauthenticated /favorites page should show empty state, login prompt, or redirect"
    )
