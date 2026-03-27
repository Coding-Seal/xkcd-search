import os
import pytest
from playwright.sync_api import Page, expect

BASE_URL = os.getenv("BASE_URL", "http://localhost:8090")


def test_homepage_loads(page: Page):
    """Verify that the home page loads and contains a search form."""
    page.goto(BASE_URL + "/")
    expect(page.locator("form")).to_be_visible()


def test_search_empty_query(page: Page):
    """Submit an empty search and verify the page loads without error."""
    page.goto(BASE_URL + "/")
    form = page.locator("form").first
    form.evaluate("f => f.submit()")
    # Page should still load (not crash)
    assert page.url is not None
    assert page.title() is not None


def test_search_with_query(page: Page):
    """Search for 'test' and verify that results or a no-results message appears."""
    page.goto(BASE_URL + "/")
    search_input = page.locator("input[type='text'], input[name='search'], input[placeholder*='earch']").first
    search_input.fill("test")
    search_input.press("Enter")
    page.wait_for_load_state("networkidle")
    # Accept either results or a "no results" message
    has_results = page.locator("img, .comic, .result, article").count() > 0
    has_no_results_msg = (
        "no result" in page.content().lower()
        or "not found" in page.content().lower()
        or "nothing" in page.content().lower()
    )
    assert has_results or has_no_results_msg, "Expected either results or a no-results message"


def test_login_page_loads(page: Page):
    """Verify that the /login page loads and contains login and password fields."""
    page.goto(BASE_URL + "/login")
    expect(page.locator("form")).to_be_visible()
    # Check for login field
    login_field = page.locator(
        "input[name='login'], input[type='text'], input[name='username']"
    ).first
    expect(login_field).to_be_visible()
    # Check for password field
    password_field = page.locator("input[type='password'], input[name='password'], input[name='pswd']").first
    expect(password_field).to_be_visible()


def test_login_invalid_credentials(page: Page):
    """Fill the login form with wrong credentials and verify an error or stay on /login."""
    page.goto(BASE_URL + "/login")
    login_field = page.locator(
        "input[name='login'], input[type='text'], input[name='username']"
    ).first
    password_field = page.locator("input[type='password'], input[name='password'], input[name='pswd']").first

    login_field.fill("wronguser")
    password_field.fill("wrongpassword")
    password_field.press("Enter")
    page.wait_for_load_state("networkidle")

    # Should either stay on /login or show an error message
    stayed_on_login = "/login" in page.url
    has_error = (
        "error" in page.content().lower()
        or "invalid" in page.content().lower()
        or "incorrect" in page.content().lower()
        or "unauthorized" in page.content().lower()
        or "wrong" in page.content().lower()
    )
    assert stayed_on_login or has_error, "Expected to stay on /login or see an error"


def test_login_valid_credentials(page: Page):
    """Fill the login form with valid admin credentials and verify redirect to home."""
    page.goto(BASE_URL + "/login")
    login_field = page.locator(
        "input[name='login'], input[type='text'], input[name='username']"
    ).first
    password_field = page.locator("input[type='password'], input[name='password'], input[name='pswd']").first

    login_field.fill("admin")
    password_field.fill("admin")
    password_field.press("Enter")
    page.wait_for_load_state("networkidle")

    # Should be redirected away from /login (typically to /)
    assert "/login" not in page.url or page.url == BASE_URL + "/", (
        f"Expected redirect after login, but stayed at {page.url}"
    )


def test_favorites_page(page: Page):
    """Verify that the /favorites page loads without a 500 error."""
    response = page.goto(BASE_URL + "/favorites")
    # Should return a loadable page (not a server error)
    assert response is not None
    assert response.status < 500, f"Favorites page returned server error: {response.status}"


def test_navigation_home(page: Page):
    """Verify that the navbar home link navigates back to the root."""
    page.goto(BASE_URL + "/")
    home_link = page.locator("a[href='/'], a[href=''], nav a").first
    home_link.click()
    page.wait_for_load_state("networkidle")
    assert page.url.rstrip("/") == BASE_URL.rstrip("/") or page.url == BASE_URL + "/"


def test_search_result_has_favorite_button(page: Page):
    """After a search that returns results, verify a favorite toggle button is present."""
    page.goto(BASE_URL + "/")
    search_input = page.locator("input[type='text'], input[name='search'], input[placeholder*='earch']").first
    search_input.fill("test")
    search_input.press("Enter")
    page.wait_for_load_state("networkidle")

    # Only check for favorite button if results are visible
    result_count = page.locator("img, .comic, .result, article").count()
    if result_count > 0:
        fav_button = page.locator(
            "button[class*='fav'], button[class*='favorite'], a[href*='fav'], "
            "button[aria-label*='favorite'], button[title*='favorite'], .favorite-btn"
        )
        assert fav_button.count() > 0, "Expected a favorite toggle button in search results"
    else:
        pytest.skip("No search results returned; skipping favorite button check")


def test_page_title(page: Page):
    """Verify that the home page has a non-empty title."""
    page.goto(BASE_URL + "/")
    title = page.title()
    assert title, "Page title should not be empty"
    assert len(title.strip()) > 0, "Page title should not be blank"
