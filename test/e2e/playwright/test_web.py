import os
from playwright.sync_api import Page, expect

BASE_URL = os.getenv("BASE_URL", "http://localhost:8090")


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
