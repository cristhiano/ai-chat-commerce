import { test, expect } from '@playwright/test';

test.describe('Chat Shopping Journey', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to chat page
    await page.goto('/chat');
    
    // Wait for chat interface to load
    await page.waitForSelector('[data-testid="chat-interface"]', { timeout: 10000 });
  });

  test('should display welcome message and connection status', async ({ page }) => {
    // Check welcome message
    await expect(page.getByText('Welcome to your shopping assistant!')).toBeVisible();
    await expect(page.getByText('Ask me about products, add items to your cart, or get recommendations.')).toBeVisible();
    
    // Check connection status
    await expect(page.getByText('Connected')).toBeVisible();
    await expect(page.locator('.bg-green-500')).toBeVisible(); // Connection indicator
  });

  test('should allow sending messages', async ({ page }) => {
    const chatInput = page.getByPlaceholderText('Ask me about products...');
    const sendButton = page.getByRole('button', { name: 'Send message' });
    
    // Type a message
    await chatInput.fill('Hello, I need help finding products');
    
    // Send message
    await sendButton.click();
    
    // Check that message appears in chat
    await expect(page.getByText('Hello, I need help finding products')).toBeVisible();
    
    // Check that input is cleared
    await expect(chatInput).toHaveValue('');
  });

  test('should handle product search through chat', async ({ page }) => {
    const chatInput = page.getByPlaceholderText('Ask me about products...');
    
    // Search for products
    await chatInput.fill('Show me wireless headphones');
    await page.getByRole('button', { name: 'Send message' }).click();
    
    // Wait for response
    await page.waitForSelector('[data-testid="assistant-message"]', { timeout: 10000 });
    
    // Check that assistant responds
    const assistantMessages = page.locator('[data-testid="assistant-message"]');
    await expect(assistantMessages.first()).toBeVisible();
  });

  test('should display product suggestions', async ({ page }) => {
    const chatInput = page.getByPlaceholderText('Ask me about products...');
    
    // Ask for product recommendations
    await chatInput.fill('What electronics do you have?');
    await page.getByRole('button', { name: 'Send message' }).click();
    
    // Wait for suggestions to appear
    await page.waitForSelector('[data-testid="product-suggestions"]', { timeout: 15000 });
    
    // Check that suggestions are displayed
    await expect(page.getByText('Product Suggestions')).toBeVisible();
    
    // Check that suggestion cards are present
    const suggestionCards = page.locator('[data-testid="suggestion-card"]');
    await expect(suggestionCards.first()).toBeVisible();
  });

  test('should allow adding products to cart through chat', async ({ page }) => {
    const chatInput = page.getByPlaceholderText('Ask me about products...');
    
    // Ask to add a product to cart
    await chatInput.fill('Add wireless headphones to my cart');
    await page.getByRole('button', { name: 'Send message' }).click();
    
    // Wait for response and action confirmation
    await page.waitForSelector('[data-testid="action-confirmation"]', { timeout: 15000 });
    
    // Check that action was taken
    await expect(page.getByText('Actions taken:')).toBeVisible();
    await expect(page.getByText('Added to cart')).toBeVisible();
  });

  test('should show cart contents when asked', async ({ page }) => {
    const chatInput = page.getByPlaceholderText('Ask me about products...');
    
    // Ask about cart contents
    await chatInput.fill('What\'s in my cart?');
    await page.getByRole('button', { name: 'Send message' }).click();
    
    // Wait for response
    await page.waitForSelector('[data-testid="assistant-message"]', { timeout: 10000 });
    
    // Check that cart information is displayed
    const assistantMessages = page.locator('[data-testid="assistant-message"]');
    await expect(assistantMessages.first()).toBeVisible();
  });

  test('should handle checkout process through chat', async ({ page }) => {
    const chatInput = page.getByPlaceholderText('Ask me about products...');
    
    // Initiate checkout
    await chatInput.fill('I want to checkout');
    await page.getByRole('button', { name: 'Send message' }).click();
    
    // Wait for response
    await page.waitForSelector('[data-testid="assistant-message"]', { timeout: 10000 });
    
    // Check that checkout guidance is provided
    const assistantMessages = page.locator('[data-testid="assistant-message"]');
    await expect(assistantMessages.first()).toBeVisible();
  });

  test('should show typing indicator when assistant is responding', async ({ page }) => {
    const chatInput = page.getByPlaceholderText('Ask me about products...');
    
    // Send a message
    await chatInput.fill('Tell me about your products');
    await page.getByRole('button', { name: 'Send message' }).click();
    
    // Check that typing indicator appears briefly
    await expect(page.getByText('Assistant is typing...')).toBeVisible({ timeout: 5000 });
    
    // Wait for response
    await page.waitForSelector('[data-testid="assistant-message"]', { timeout: 15000 });
    
    // Check that typing indicator disappears
    await expect(page.getByText('Assistant is typing...')).not.toBeVisible();
  });

  test('should handle connection errors gracefully', async ({ page }) => {
    // Simulate connection error by navigating away and back
    await page.goto('/');
    await page.goto('/chat');
    
    // Check that error handling is in place
    await page.waitForSelector('[data-testid="chat-interface"]', { timeout: 10000 });
    
    // Should still be able to use the interface
    const chatInput = page.getByPlaceholderText('Ask me about products...');
    await expect(chatInput).toBeVisible();
  });

  test('should maintain conversation history', async ({ page }) => {
    const chatInput = page.getByPlaceholderText('Ask me about products...');
    
    // Send multiple messages
    await chatInput.fill('Hello');
    await page.getByRole('button', { name: 'Send message' }).click();
    
    await page.waitForSelector('[data-testid="assistant-message"]', { timeout: 10000 });
    
    await chatInput.fill('What products do you have?');
    await page.getByRole('button', { name: 'Send message' }).click();
    
    await page.waitForSelector('[data-testid="assistant-message"]', { timeout: 10000 });
    
    // Check that both messages are visible
    await expect(page.getByText('Hello')).toBeVisible();
    await expect(page.getByText('What products do you have?')).toBeVisible();
  });

  test('should allow clicking on product suggestions', async ({ page }) => {
    const chatInput = page.getByPlaceholderText('Ask me about products...');
    
    // Ask for product recommendations
    await chatInput.fill('Show me electronics');
    await page.getByRole('button', { name: 'Send message' }).click();
    
    // Wait for suggestions
    await page.waitForSelector('[data-testid="suggestion-card"]', { timeout: 15000 });
    
    // Click on first suggestion
    const firstSuggestion = page.locator('[data-testid="suggestion-card"]').first();
    await firstSuggestion.click();
    
    // Check that a new message was sent
    await expect(page.getByText('Tell me more about')).toBeVisible();
  });

  test('should display usage instructions', async ({ page }) => {
    // Check that usage instructions are visible
    await expect(page.getByText('Ask about products: "Show me wireless headphones"')).toBeVisible();
    await expect(page.getByText('Add to cart: "Add the blue t-shirt to my cart"')).toBeVisible();
    await expect(page.getByText('Complete purchase: "I want to checkout"')).toBeVisible();
  });

  test('should handle empty messages gracefully', async ({ page }) => {
    const sendButton = page.getByRole('button', { name: 'Send message' });
    
    // Try to send empty message
    await sendButton.click();
    
    // Should not send anything
    await expect(page.getByText('Welcome to your shopping assistant!')).toBeVisible();
  });

  test('should handle long messages', async ({ page }) => {
    const chatInput = page.getByPlaceholderText('Ask me about products...');
    const longMessage = 'This is a very long message that should be handled properly by the chat interface. '.repeat(10);
    
    // Type long message
    await chatInput.fill(longMessage);
    
    // Check character count
    await expect(page.getByText('500/500')).toBeVisible();
    
    // Send message
    await page.getByRole('button', { name: 'Send message' }).click();
    
    // Should handle the message
    await expect(page.getByText(longMessage.substring(0, 50))).toBeVisible();
  });
});
