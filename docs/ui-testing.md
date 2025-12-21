# UI Testing DSL Specification

## Overview

High-level, intent-based UI testing for frontend applications. Describe **what** should happen, not **how**.

## File Format

UI tests use the same Orchestrfile format with `TYPE: UI` header:

```
TYPE: UI
BASE: http://localhost:5173

# Test cases follow...
```

## Keywords

### Header Keywords

| Keyword | Description | Example |
|---------|-------------|---------|
| `TYPE` | Test type (required for UI) | `TYPE: UI` |
| `BASE` | Base URL for the app | `BASE: http://localhost:5173` |

### Test Case Keywords

| Keyword | Description | Example |
|---------|-------------|---------|
| `FLOW` | User flow name | `FLOW: Login` |
| `PAGE` | Starting page (optional) | `PAGE: /login` |
| `DATA` | Form input data | `DATA: {email: "test@test.com"}` |
| `EXPECT` | Expected outcome | `EXPECT: /dashboard` |
| `EXPECT_TEXT` | Expected text on page | `EXPECT_TEXT: "Welcome"` |
| `EXPECT_ELEMENT` | Expected element exists | `EXPECT_ELEMENT: .success-message` |
| `EXPECT_COUNT` | Element count assertion | `EXPECT_COUNT: .item >= 1` |
| `EXPECT_URL` | Expected URL pattern | `EXPECT_URL: /users/*` |

## Common Flows

The tool recognizes common flow patterns and auto-navigates:

| Flow Name | Auto-navigates to | Looks for |
|-----------|-------------------|-----------|
| `Login` | `/login`, `/signin` | Login form |
| `Signup` | `/signup`, `/register` | Registration form |
| `Logout` | Current page | Logout button/link |
| `Search` | Current page | Search input |
| `Add to cart` | Product page | Add to cart button |
| `Checkout` | `/cart`, `/checkout` | Checkout form |

For custom flows, use `PAGE:` to specify the starting point.

## Examples

### Basic Login Flow

```
TYPE: UI
BASE: http://localhost:5173

FLOW: Login
DATA: {email: "test@example.com", password: "test123"}
EXPECT: /dashboard
EXPECT_TEXT: "Welcome"
```

### Signup with Validation

```
TYPE: UI
BASE: http://localhost:5173

FLOW: Signup
DATA: {
  name: "John Doe",
  email: "john@example.com",
  password: "securepass123",
  confirm_password: "securepass123"
}
EXPECT: /welcome
EXPECT_TEXT: "Welcome, John"
```

### E-commerce Flow

```
TYPE: UI
BASE: http://localhost:5173

FLOW: Search
DATA: {query: "laptop"}
EXPECT_COUNT: .product-card >= 1

---

FLOW: Add to cart
PAGE: /products/1
EXPECT_TEXT: "Added to cart"

---

FLOW: Checkout
DATA: {
  address: "123 Main St",
  city: "New York",
  zip: "10001",
  card: "4242424242424242"
}
EXPECT: /order-confirmation
EXPECT_TEXT: "Order placed"
```

### Custom Flow

```
TYPE: UI
BASE: http://localhost:5173

FLOW: Update profile
PAGE: /settings/profile
DATA: {name: "Jane Doe", bio: "Developer"}
EXPECT_TEXT: "Profile updated"
```

### Page Smoke Tests

```
TYPE: UI
BASE: http://localhost:5173

PAGE: /
EXPECT: 200
EXPECT_ELEMENT: header
EXPECT_ELEMENT: nav
EXPECT_ELEMENT: footer

---

PAGE: /about
EXPECT: 200
EXPECT_TEXT: "About Us"

---

PAGE: /contact
EXPECT: 200
EXPECT_ELEMENT: form
```

## How It Works

### Form Detection

The tool automatically:

1. **Finds forms** on the page
2. **Matches DATA fields** to inputs by:
   - `name` attribute
   - `id` attribute
   - `label` text
   - `placeholder` text
3. **Fills inputs** with provided values
4. **Finds submit button** and clicks it

### Field Matching Examples

```
DATA: {email: "test@test.com"}
```

Matches any of:
- `<input name="email">`
- `<input id="email">`
- `<label>Email</label> <input>`
- `<input placeholder="Email">`

### Outcome Verification

| EXPECT Format | Verification |
|---------------|--------------|
| `/path` | URL ends with path |
| `/path/*` | URL matches pattern |
| `200` | HTTP status code |
| Text in quotes | Text visible on page |

## Comparison: API vs UI

| Aspect | API Testing | UI Testing |
|--------|-------------|------------|
| Target | HTTP endpoints | Browser pages |
| Input | JSON request body | Form fields |
| Output | JSON response | Page state |
| Assertions | Status + body pattern | URL + text + elements |

### Same Orchestrfile, Different Types

**API Test:**
```
TYPE: API

ENDPOINT: /api/login
METHOD: POST
DATA: {email: "test@test.com", password: "test123"}
EXPECT: 200 with {token: <jwt>}
```

**UI Test:**
```
TYPE: UI

FLOW: Login
DATA: {email: "test@test.com", password: "test123"}
EXPECT: /dashboard
```

## Implementation Notes

### Browser Automation

Uses headless browser (Chromedp/Playwright) under the hood:

1. Launch headless browser
2. Navigate to BASE + PAGE (or auto-detected path)
3. Wait for page load
4. Find and fill form fields
5. Click submit
6. Wait for navigation/response
7. Verify expectations

### Error Handling

| Error | Report |
|-------|--------|
| Form not found | "No form found on /login" |
| Field not found | "Could not find input for 'email'" |
| Submit failed | "No submit button found" |
| Timeout | "Page did not respond within 30s" |
| Assertion failed | "Expected '/dashboard', got '/login'" |

### Timeouts

- Page load: 30s
- Form fill: 5s per field
- Navigation: 30s
- Element wait: 10s

## Future Enhancements

- **Visual regression** - Screenshot comparison
- **Network mocking** - Mock API responses during UI tests
- **Multi-step flows** - Chain multiple actions
- **Conditional logic** - If/else in flows
- **Data generation** - Random emails, names, etc.
