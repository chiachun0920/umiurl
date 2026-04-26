Feature: User create short url

    Scenario: User create short url successfully
        When I create a short url with "https://www.example.com"
        Then I should receive a short url
        And the short url should redirect to "https://www.example.com"

    Scenario: User create short url with invalid URL
        When I create a short url with "invalid-url"
        Then I should receive an error message "url must be absolute http or https"