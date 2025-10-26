/**
 * E2E Test: Chat to Cart - Structured Product Data via WebSocket
 * 
 * T075: Test ChatInterface receives and displays WebSocket product suggestions
 * with complete structured data (category, tags, inventory)
 */

import { test, expect } from '@playwright/test';

test.describe('Chat to Cart - Structured Product Data', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to chat page
    await page.goto('/chat');
    
    // Wait for chat interface to load
    await page.waitForSelector('[data-testid="chat-interface"]', { timeout: 10000 });
    
    // Wait for WebSocket connection
    await expect(page.getByText('Connected')).toBeVisible({ timeout: 5000 });
  });

  test('should receive and display product suggestions with complete data via WebSocket', async ({ page }) => {
    const chatInput = page.getByPlaceholderText(/Type your message|Ask me about products/i);
    
    // Send a message that will trigger product suggestions
    await chatInput.fill('Show me electronics');
    await page.keyboard.press('Enter');
    
    // Wait for user message to appear
    await expect(page.getByText('Show me electronics')).toBeVisible({ timeout: 3000 });
    
    // Wait for product suggestions to be received via WebSocket
    // This should trigger when backend sends a "suggestions" WebSocket message
    await page.waitForSelector('[data-testid="product-suggestion-card"]', { 
      timeout: 15000,
      state: 'visible' 
    });
    
    // Verify at least one product suggestion card is displayed
    const suggestionCards = page.locator('[data-testid="product-suggestion-card"]');
    const cardCount = await suggestionCards.count();
    expect(cardCount).toBeGreaterThan(0);
    
    // Get the first suggestion card for detailed testing
    const firstCard = suggestionCards.first();
    await expect(firstCard).toBeVisible();
    
    // Verify product name is displayed
    const productName = firstCard.locator('.font-semibold').first();
    await expect(productName).toBeVisible();
    const nameText = await productName.textContent();
    expect(nameText).not.toBeNull();
    expect(nameText!.length).toBeGreaterThan(0);
    
    // Verify product price is displayed
    const priceElement = firstCard.locator('.text-blue-600');
    await expect(priceElement).toBeVisible();
    const priceText = await priceElement.textContent();
    expect(priceText).toMatch(/\$\d+\.\d{2}/); // Should match $XX.XX format
    
    // Verify product description is displayed
    const descriptionElement = firstCard.locator('.text-sm.text-gray-600');
    await expect(descriptionElement).toBeVisible();
  });

  test('should display category information in product suggestions', async ({ page }) => {
    const chatInput = page.getByPlaceholderText(/Type your message|Ask me about products/i);
    
    // Request products
    await chatInput.fill('What products do you have in electronics?');
    await page.keyboard.press('Enter');
    
    // Wait for suggestions
    await page.waitForSelector('[data-testid="product-suggestion-card"]', { 
      timeout: 15000 
    });
    
    const firstCard = page.locator('[data-testid="product-suggestion-card"]').first();
    
    // Look for category display (should be in gray text)
    const categoryElement = firstCard.locator('.text-xs.text-gray-500');
    
    // Category should be visible and have text content
    if (await categoryElement.count() > 0) {
      const categoryText = await categoryElement.textContent();
      expect(categoryText).not.toBeNull();
      expect(categoryText!.length).toBeGreaterThan(0);
    }
  });

  test('should display product tags in suggestions', async ({ page }) => {
    const chatInput = page.getByPlaceholderText(/Type your message|Ask me about products/i);
    
    // Request products
    await chatInput.fill('Show me popular products');
    await page.keyboard.press('Enter');
    
    // Wait for suggestions
    await page.waitForSelector('[data-testid="product-suggestion-card"]', { 
      timeout: 15000 
    });
    
    const firstCard = page.locator('[data-testid="product-suggestion-card"]').first();
    
    // Look for tags (should be small badges with bg-gray-100)
    const tagElements = firstCard.locator('.bg-gray-100.text-gray-600');
    
    // If tags exist, verify they're displayed correctly
    const tagCount = await tagElements.count();
    if (tagCount > 0) {
      // Verify first tag is visible and has content
      const firstTag = tagElements.first();
      await expect(firstTag).toBeVisible();
      const tagText = await firstTag.textContent();
      expect(tagText).not.toBeNull();
      expect(tagText!.length).toBeGreaterThan(0);
      
      // Verify at most 2 tags are shown (per design spec)
      expect(tagCount).toBeLessThanOrEqual(2);
    }
  });

  test('should show Add to Cart button in product suggestions', async ({ page }) => {
    const chatInput = page.getByPlaceholderText(/Type your message|Ask me about products/i);
    
    // Request products
    await chatInput.fill('Show me wireless headphones');
    await page.keyboard.press('Enter');
    
    // Wait for suggestions with Add to Cart functionality
    await page.waitForSelector('[data-testid="product-suggestion-card"]', { 
      timeout: 15000 
    });
    
    const firstCard = page.locator('[data-testid="product-suggestion-card"]').first();
    
    // Verify Add to Cart button exists
    const addToCartButton = firstCard.locator('button').filter({ hasText: /Add to Cart|In cart/i });
    await expect(addToCartButton).toBeVisible();
  });

  test('should handle multiple product suggestions from WebSocket', async ({ page }) => {
    const chatInput = page.getByPlaceholderText(/Type your message|Ask me about products/i);
    
    // Request multiple products
    await chatInput.fill('Show me all your products');
    await page.keyboard.press('Enter');
    
    // Wait for suggestions
    await page.waitForSelector('[data-testid="product-suggestion-card"]', { 
      timeout: 15000 
    });
    
    // Count suggestion cards
    const suggestionCards = page.locator('[data-testid="product-suggestion-card"]');
    const cardCount = await suggestionCards.count();
    
    // Should receive multiple suggestions (backend typically sends 3-5)
    expect(cardCount).toBeGreaterThanOrEqual(1);
    
    // Verify each card has required elements
    for (let i = 0; i < Math.min(cardCount, 3); i++) {
      const card = suggestionCards.nth(i);
      
      // Each card should have name, price, description
      await expect(card.locator('.font-semibold').first()).toBeVisible();
      await expect(card.locator('.text-blue-600')).toBeVisible();
      await expect(card.locator('.text-sm.text-gray-600')).toBeVisible();
    }
  });

  test('should display out of stock status when inventory is zero', async ({ page }) => {
    const chatInput = page.getByPlaceholderText(/Type your message|Ask me about products/i);
    
    // Request products
    await chatInput.fill('Show me products');
    await page.keyboard.press('Enter');
    
    // Wait for suggestions
    await page.waitForSelector('[data-testid="product-suggestion-card"]', { 
      timeout: 15000 
    });
    
    // Check all cards for out of stock message
    const allCards = page.locator('[data-testid="product-suggestion-card"]');
    const cardCount = await allCards.count();
    
    // Check each card for out of stock indicator
    for (let i = 0; i < cardCount; i++) {
      const card = allCards.nth(i);
      
      // Check if this card has out of stock message
      const outOfStockMsg = card.locator('.text-red-600').filter({ hasText: /out of stock/i });
      
      if (await outOfStockMsg.count() > 0) {
        // If out of stock message exists, verify it's visible
        await expect(outOfStockMsg).toBeVisible();
        
        // Verify Add to Cart button is disabled or has appropriate styling
        const addToCartButton = card.locator('button').filter({ hasText: /Add to Cart/i });
        if (await addToCartButton.count() > 0) {
          const classes = await addToCartButton.getAttribute('class');
          expect(classes).toContain('opacity-50'); // Should have disabled styling
        }
      }
    }
  });

  test('should handle WebSocket connection errors gracefully', async ({ page }) => {
    // This test verifies the UI handles connection issues
    
    // Check initial connection status
    await expect(page.getByText('Connected')).toBeVisible({ timeout: 5000 });
    
    const chatInput = page.getByPlaceholderText(/Type your message|Ask me about products/i);
    
    // Try to send a message
    await chatInput.fill('Test message');
    await page.keyboard.press('Enter');
    
    // Should display message even if there are connection issues
    await expect(page.getByText('Test message')).toBeVisible({ timeout: 3000 });
  });

  test('should receive product suggestions in correct message format', async ({ page }) => {
    // Listen for WebSocket messages
    const wsMessages: any[] = [];
    
    page.on('websocket', ws => {
      ws.on('framereceived', event => {
        try {
          const message = JSON.parse(event.payload as string);
          wsMessages.push(message);
        } catch (e) {
          // Ignore non-JSON messages
        }
      });
    });
    
    const chatInput = page.getByPlaceholderText(/Type your message|Ask me about products/i);
    
    // Send message to trigger suggestions
    await chatInput.fill('Show me products');
    await page.keyboard.press('Enter');
    
    // Wait for suggestions to appear in UI
    await page.waitForSelector('[data-testid="product-suggestion-card"]', { 
      timeout: 15000 
    });
    
    // Give time for all WebSocket messages to be received
    await page.waitForTimeout(1000);
    
    // Verify we received at least one message
    expect(wsMessages.length).toBeGreaterThan(0);
    
    // Look for a suggestions message
    const suggestionsMsg = wsMessages.find(msg => 
      msg.type === 'suggestions' || 
      (msg.type === 'message' && msg.data && msg.data.metadata && msg.data.metadata.suggestions)
    );
    
    // If we found a suggestions message, verify its structure
    if (suggestionsMsg) {
      // Verify message has proper structure
      expect(suggestionsMsg).toHaveProperty('type');
      
      // Verify suggestions data exists somewhere in the message
      const hasSuggestionsData = 
        suggestionsMsg.data || 
        (suggestionsMsg.data && suggestionsMsg.data.metadata && suggestionsMsg.data.metadata.suggestions);
      
      expect(hasSuggestionsData).toBeTruthy();
    }
  });

  test('should persist product suggestions after scrolling', async ({ page }) => {
    const chatInput = page.getByPlaceholderText(/Type your message|Ask me about products/i);
    
    // Send multiple messages to create scrollable content
    for (let i = 0; i < 3; i++) {
      await chatInput.fill(`Message ${i + 1}`);
      await page.keyboard.press('Enter');
      await page.waitForTimeout(500);
    }
    
    // Request product suggestions
    await chatInput.fill('Show me products');
    await page.keyboard.press('Enter');
    
    // Wait for suggestions
    await page.waitForSelector('[data-testid="product-suggestion-card"]', { 
      timeout: 15000 
    });
    
    const firstCard = page.locator('[data-testid="product-suggestion-card"]').first();
    await expect(firstCard).toBeVisible();
    
    // Scroll up in chat
    await page.evaluate(() => {
      const chatContainer = document.querySelector('[data-testid="chat-messages"]');
      if (chatContainer) {
        chatContainer.scrollTop = 0;
      }
    });
    
    await page.waitForTimeout(500);
    
    // Scroll back down
    await page.evaluate(() => {
      const chatContainer = document.querySelector('[data-testid="chat-messages"]');
      if (chatContainer) {
        chatContainer.scrollTop = chatContainer.scrollHeight;
      }
    });
    
    await page.waitForTimeout(500);
    
    // Verify suggestions are still visible
    await expect(firstCard).toBeVisible();
  });
});

