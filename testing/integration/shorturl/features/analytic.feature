@debug
Feature: Analytics for short url

Scenario: User view short url analytics
    Given a short url exists for "https://www.example.com" with code "abc123"
    When I view the analytics for the short url with code "abc123"
    Then I should receive analytics data for the short url with code "abc123"
    When I visit the short url with code "abc123"
    When I view the analytics for the short url with code "abc123"
    Then I should receive analytics data for the short url with code "abc123" with 1 visit