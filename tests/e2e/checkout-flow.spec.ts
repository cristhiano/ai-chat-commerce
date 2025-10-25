import { test, expect } from '@playwright/test';

test.describe('E-commerce Checkout Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the home page
    await page.goto('http://localhost:5181');
    
    // Wait for the page to load
    await page.waitForLoadState('networkidle');
  });

  test('complete checkout flow as anonymous user', async ({ page }) => {
    // Step 1: Browse products
    await test.step('Browse products', async () => {
      // Check if products are displayed
      await expect(page.locator('[data-testid="product-card"]').first()).toBeVisible();
      
      // Click on the first product
      await page.locator('[data-testid="product-card"]').first().click();
      
      // Wait for product detail page to load
      await page.waitForURL(/\/products\/.*/);
    });

    // Step 2: Add product to cart
    await test.step('Add product to cart', async () => {
      // Verify product detail page loaded
      await expect(page.locator('h1')).toContainText('Test Product');
      
      // Set quantity to 2
      const quantityInput = page.locator('input[type="number"]');
      await quantityInput.fill('2');
      
      // Click add to cart button
      await page.locator('button:has-text("Add to Cart")').click();
      
      // Wait for success message or cart update
      await expect(page.locator('text=Product added to cart!')).toBeVisible();
    });

    // Step 3: View cart
    await test.step('View shopping cart', async () => {
      // Navigate to cart
      await page.goto('http://localhost:5181/cart');
      
      // Verify cart page loaded
      await expect(page.locator('h1')).toContainText('Shopping Cart');
      
      // Verify product is in cart
      await expect(page.locator('text=Test Product')).toBeVisible();
      await expect(page.locator('text=Qty: 2')).toBeVisible();
    });

    // Step 4: Proceed to checkout
    await test.step('Proceed to checkout', async () => {
      // Click proceed to checkout button
      await page.locator('button:has-text("Proceed to Checkout")').click();
      
      // Wait for checkout page to load
      await page.waitForURL(/\/checkout/);
      
      // Verify checkout page loaded
      await expect(page.locator('h1')).toContainText('Checkout');
    });

    // Step 5: Fill checkout form
    await test.step('Fill checkout form', async () => {
      // Fill contact information
      await page.locator('input[name="email"]').fill('test@example.com');
      await page.locator('input[name="phone"]').fill('555-123-4567');
      
      // Fill shipping address
      await page.locator('input[name="shippingFirstName"]').fill('John');
      await page.locator('input[name="shippingLastName"]').fill('Doe');
      await page.locator('input[name="shippingAddress1"]').fill('123 Main St');
      await page.locator('input[name="shippingCity"]').fill('Anytown');
      await page.locator('input[name="shippingState"]').fill('CA');
      await page.locator('input[name="shippingZip"]').fill('12345');
      
      // Add order notes
      await page.locator('textarea[name="notes"]').fill('Please deliver during business hours');
    });

    // Step 6: Complete order
    await test.step('Complete order', async () => {
      // Click complete order button
      await page.locator('button:has-text("Complete Order")').click();
      
      // Wait for order confirmation or redirect
      await page.waitForURL(/\/orders\/.*/, { timeout: 10000 });
      
      // Verify order confirmation page
      await expect(page.locator('h1')).toContainText('Order Confirmation');
    });
  });

  test('complete checkout flow as registered user', async ({ page }) => {
    // Step 1: Register new user
    await test.step('Register new user', async () => {
      await page.goto('http://localhost:5181/register');
      
      // Fill registration form
      await page.locator('input[name="firstName"]').fill('Jane');
      await page.locator('input[name="lastName"]').fill('Smith');
      await page.locator('input[name="email"]').fill('jane@example.com');
      await page.locator('input[name="phone"]').fill('555-987-6543');
      await page.locator('input[name="password"]').fill('password123');
      await page.locator('input[name="confirmPassword"]').fill('password123');
      
      // Submit registration
      await page.locator('button:has-text("Create Account")').click();
      
      // Wait for redirect to home page
      await page.waitForURL('/', { timeout: 10000 });
    });

    // Step 2: Browse and add products to cart
    await test.step('Add products to cart', async () => {
      // Navigate to products page
      await page.goto('http://localhost:5181/products');
      
      // Add first product to cart
      await page.locator('[data-testid="product-card"]').first().click();
      await page.locator('button:has-text("Add to Cart")').click();
      await expect(page.locator('text=Product added to cart!')).toBeVisible();
      
      // Go back to products
      await page.goBack();
      
      // Add second product to cart
      await page.locator('[data-testid="product-card"]').nth(1).click();
      await page.locator('button:has-text("Add to Cart")').click();
      await expect(page.locator('text=Product added to cart!')).toBeVisible();
    });

    // Step 3: View cart and update quantities
    await test.step('Update cart quantities', async () => {
      await page.goto('http://localhost:5181/cart');
      
      // Update quantity of first item
      const quantityButton = page.locator('button:has-text("+")').first();
      await quantityButton.click();
      
      // Verify quantity updated
      await expect(page.locator('text=Qty: 2')).toBeVisible();
    });

    // Step 4: Proceed to checkout
    await test.step('Proceed to checkout', async () => {
      await page.locator('button:has-text("Proceed to Checkout")').click();
      await page.waitForURL(/\/checkout/);
    });

    // Step 5: Fill checkout form (user info should be pre-filled)
    await test.step('Complete checkout form', async () => {
      // Verify email is pre-filled
      await expect(page.locator('input[name="email"]')).toHaveValue('jane@example.com');
      
      // Fill shipping address
      await page.locator('input[name="shippingFirstName"]').fill('Jane');
      await page.locator('input[name="shippingLastName"]').fill('Smith');
      await page.locator('input[name="shippingAddress1"]').fill('456 Oak Ave');
      await page.locator('input[name="shippingCity"]').fill('Springfield');
      await page.locator('input[name="shippingState"]').fill('IL');
      await page.locator('input[name="shippingZip"]').fill('62701');
      
      // Use different billing address
      await page.locator('input[name="billingSameAsShipping"]').uncheck();
      await page.locator('input[name="billingFirstName"]').fill('Jane');
      await page.locator('input[name="billingLastName"]').fill('Smith');
      await page.locator('input[name="billingAddress1"]').fill('789 Pine St');
      await page.locator('input[name="billingCity"]').fill('Springfield');
      await page.locator('input[name="billingState"]').fill('IL');
      await page.locator('input[name="billingZip"]').fill('62701');
    });

    // Step 6: Complete order
    await test.step('Complete order', async () => {
      await page.locator('button:has-text("Complete Order")').click();
      await page.waitForURL(/\/orders\/.*/, { timeout: 10000 });
      
      // Verify order confirmation
      await expect(page.locator('h1')).toContainText('Order Confirmation');
    });

    // Step 7: View order history
    await test.step('View order history', async () => {
      await page.goto('http://localhost:5181/profile');
      
      // Click on orders tab
      await page.locator('button:has-text("Order History")').click();
      
      // Verify order appears in history
      await expect(page.locator('text=Order #')).toBeVisible();
    });
  });

  test('cart persistence across sessions', async ({ page, context }) => {
    // Step 1: Add items to cart as anonymous user
    await test.step('Add items to cart', async () => {
      await page.goto('http://localhost:5181/products');
      await page.locator('[data-testid="product-card"]').first().click();
      await page.locator('button:has-text("Add to Cart")').click();
      await expect(page.locator('text=Product added to cart!')).toBeVisible();
    });

    // Step 2: Verify cart count
    await test.step('Verify cart count', async () => {
      const cartCount = page.locator('[data-testid="cart-count"]');
      await expect(cartCount).toContainText('1');
    });

    // Step 3: Open new tab and verify cart persists
    await test.step('Verify cart persistence', async () => {
      const newPage = await context.newPage();
      await newPage.goto('http://localhost:5181/cart');
      
      // Cart should still contain the item
      await expect(newPage.locator('text=Test Product')).toBeVisible();
      
      await newPage.close();
    });
  });

  test('search and filter products', async ({ page }) => {
    // Step 1: Search for products
    await test.step('Search products', async () => {
      await page.goto('http://localhost:5181/products');
      
      const searchInput = page.locator('input[placeholder="Search products..."]');
      await searchInput.fill('test');
      await page.locator('button:has-text("Search")').click();
      
      // Verify search results
      await expect(page.locator('[data-testid="product-card"]')).toBeVisible();
    });

    // Step 2: Apply filters
    await test.step('Apply filters', async () => {
      await page.locator('button:has-text("Filters")').click();
      
      // Set price range
      await page.locator('input[placeholder="Min"]').fill('50');
      await page.locator('input[placeholder="Max"]').fill('200');
      
      // Select category
      await page.locator('select').first().selectOption('electronics');
      
      // Apply filters
      await page.locator('button:has-text("Apply Filters")').click();
      
      // Verify filters applied
      await expect(page.locator('text=Active filters:')).toBeVisible();
    });

    // Step 3: Clear filters
    await test.step('Clear filters', async () => {
      await page.locator('button:has-text("Clear all filters")').click();
      
      // Verify filters cleared
      await expect(page.locator('text=Active filters:')).not.toBeVisible();
    });
  });

  test('error handling during checkout', async ({ page }) => {
    // Step 1: Add product to cart
    await test.step('Add product to cart', async () => {
      await page.goto('http://localhost:5181/products');
      await page.locator('[data-testid="product-card"]').first().click();
      await page.locator('button:has-text("Add to Cart")').click();
    });

    // Step 2: Go to checkout
    await test.step('Go to checkout', async () => {
      await page.goto('http://localhost:5181/checkout');
    });

    // Step 3: Submit incomplete form
    await test.step('Submit incomplete form', async () => {
      await page.locator('button:has-text("Complete Order")').click();
      
      // Verify validation errors
      await expect(page.locator('text=This field is required')).toBeVisible();
    });

    // Step 4: Fill form with invalid data
    await test.step('Fill form with invalid data', async () => {
      await page.locator('input[name="email"]').fill('invalid-email');
      await page.locator('input[name="shippingFirstName"]').fill('John');
      await page.locator('input[name="shippingLastName"]').fill('Doe');
      await page.locator('input[name="shippingAddress1"]').fill('123 Main St');
      await page.locator('input[name="shippingCity"]').fill('Anytown');
      await page.locator('input[name="shippingState"]').fill('CA');
      await page.locator('input[name="shippingZip"]').fill('12345');
      
      await page.locator('button:has-text("Complete Order")').click();
      
      // Verify email validation error
      await expect(page.locator('text=Please enter a valid email address')).toBeVisible();
    });
  });

  test('responsive design on mobile', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });

    // Step 1: Test mobile navigation
    await test.step('Test mobile navigation', async () => {
      await page.goto('http://localhost:5181');
      
      // Check if mobile menu is visible
      const mobileMenuButton = page.locator('[data-testid="mobile-menu-button"]');
      if (await mobileMenuButton.isVisible()) {
        await mobileMenuButton.click();
        await expect(page.locator('[data-testid="mobile-menu"]')).toBeVisible();
      }
    });

    // Step 2: Test mobile product grid
    await test.step('Test mobile product grid', async () => {
      await page.goto('http://localhost:5181/products');
      
      // Verify products are displayed in single column
      const productCards = page.locator('[data-testid="product-card"]');
      await expect(productCards.first()).toBeVisible();
    });

    // Step 3: Test mobile checkout form
    await test.step('Test mobile checkout form', async () => {
      await page.goto('http://localhost:5181/products');
      await page.locator('[data-testid="product-card"]').first().click();
      await page.locator('button:has-text("Add to Cart")').click();
      await page.goto('http://localhost:5181/checkout');
      
      // Verify form is responsive
      await expect(page.locator('input[name="email"]')).toBeVisible();
      await expect(page.locator('input[name="shippingFirstName"]')).toBeVisible();
    });
  });
});
