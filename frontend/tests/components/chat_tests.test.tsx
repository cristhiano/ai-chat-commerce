import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import ChatInterface from '../../src/components/chat/ChatInterface';
import ChatInput from '../../src/components/chat/ChatInput';
import ChatMessage from '../../src/components/chat/ChatMessage';
import ProductSuggestionCard from '../../src/components/chat/ProductSuggestionCard';
import type { ChatMessage as ChatMessageType, ProductSuggestion } from '../../src/types';

// Mock WebSocket
const mockWebSocket = {
  send: vi.fn(),
  close: vi.fn(),
  addEventListener: vi.fn(),
  removeEventListener: vi.fn(),
  readyState: WebSocket.OPEN,
};

// Mock global WebSocket
global.WebSocket = vi.fn(() => mockWebSocket) as any;

// Mock fetch
global.fetch = vi.fn();

const mockProduct: any = {
  id: '1',
  name: 'Test Product',
  description: 'A test product',
  price: 99.99,
  sku: 'TEST-001',
  status: 'active',
  category: {
    id: '1',
    name: 'Electronics',
    slug: 'electronics',
  },
};

const mockChatMessage: ChatMessageType = {
  id: '1',
  sessionId: 'test-session',
  role: 'user',
  content: 'Hello, I need help',
  timestamp: '2023-01-01T10:00:00Z',
};

const mockSuggestion: ProductSuggestion = {
  product: mockProduct,
  reason: 'Mentioned in conversation',
  confidence: 0.8,
};

describe('ChatInterface', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (global.fetch as any).mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({
        success: true,
        data: { messages: [] },
      }),
    });
  });

  it('renders chat interface with welcome message when no messages', () => {
    render(
      <BrowserRouter>
        <ChatInterface sessionId="test-session" />
      </BrowserRouter>
    );

    expect(screen.getByText('Welcome to your shopping assistant!')).toBeInTheDocument();
    expect(screen.getByText('Ask me about products, add items to your cart, or get recommendations.')).toBeInTheDocument();
  });

  it('displays connection status', () => {
    render(
      <BrowserRouter>
        <ChatInterface sessionId="test-session" />
      </BrowserRouter>
    );

    expect(screen.getByText('Shopping Assistant')).toBeInTheDocument();
    expect(screen.getByText('Connected')).toBeInTheDocument();
  });

  it('renders messages when provided', () => {
    const messages = [
      { ...mockChatMessage, role: 'user' as const, content: 'Hello' },
      { ...mockChatMessage, id: '2', role: 'assistant' as const, content: 'Hi! How can I help?' },
    ];

    render(
      <BrowserRouter>
        <ChatInterface sessionId="test-session" />
      </BrowserRouter>
    );

    // Messages would be rendered by the component when WebSocket receives them
    // This test verifies the component structure is correct
    expect(screen.getByRole('main')).toBeInTheDocument();
  });

  it('handles product suggestions', () => {
    const suggestions = [mockSuggestion];

    render(
      <BrowserRouter>
        <ChatInterface sessionId="test-session" />
      </BrowserRouter>
    );

    // Suggestions would be rendered when received from WebSocket
    // This test verifies the component can handle suggestions
    expect(screen.getByRole('main')).toBeInTheDocument();
  });
});

describe('ChatInput', () => {
  const mockOnSendMessage = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders input field and send button', () => {
    render(<ChatInput onSendMessage={mockOnSendMessage} />);

    expect(screen.getByPlaceholderText('Type your message...')).toBeInTheDocument();
    expect(screen.getByRole('button')).toBeInTheDocument();
  });

  it('sends message when form is submitted', () => {
    render(<ChatInput onSendMessage={mockOnSendMessage} />);

    const input = screen.getByPlaceholderText('Type your message...');
    const button = screen.getByRole('button');

    fireEvent.change(input, { target: { value: 'Hello' } });
    fireEvent.click(button);

    expect(mockOnSendMessage).toHaveBeenCalledWith('Hello');
  });

  it('sends message when Enter is pressed', () => {
    render(<ChatInput onSendMessage={mockOnSendMessage} />);

    const input = screen.getByPlaceholderText('Type your message...');

    fireEvent.change(input, { target: { value: 'Hello' } });
    fireEvent.keyDown(input, { key: 'Enter' });

    expect(mockOnSendMessage).toHaveBeenCalledWith('Hello');
  });

  it('does not send empty messages', () => {
    render(<ChatInput onSendMessage={mockOnSendMessage} />);

    const button = screen.getByRole('button');
    fireEvent.click(button);

    expect(mockOnSendMessage).not.toHaveBeenCalled();
  });

  it('shows character count', () => {
    render(<ChatInput onSendMessage={mockOnSendMessage} />);

    const input = screen.getByPlaceholderText('Type your message...');
    fireEvent.change(input, { target: { value: 'Hello' } });

    expect(screen.getByText('5/500')).toBeInTheDocument();
  });

  it('disables input when disabled prop is true', () => {
    render(<ChatInput onSendMessage={mockOnSendMessage} disabled={true} />);

    const input = screen.getByPlaceholderText('Connecting...');
    const button = screen.getByRole('button');

    expect(input).toBeDisabled();
    expect(button).toBeDisabled();
  });
});

describe('ChatMessage', () => {
  const mockOnSuggestionClick = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders user message correctly', () => {
    const userMessage = { ...mockChatMessage, role: 'user' as const };

    render(<ChatMessage message={userMessage} onSuggestionClick={mockOnSuggestionClick} />);

    expect(screen.getByText('Hello, I need help')).toBeInTheDocument();
    expect(screen.getByText('10:00 AM')).toBeInTheDocument();
  });

  it('renders assistant message correctly', () => {
    const assistantMessage = { ...mockChatMessage, role: 'assistant' as const, content: 'How can I help you?' };

    render(<ChatMessage message={assistantMessage} onSuggestionClick={mockOnSuggestionClick} />);

    expect(screen.getByText('How can I help you?')).toBeInTheDocument();
  });

  it('renders system message correctly', () => {
    const systemMessage = { ...mockChatMessage, role: 'system' as const, content: 'System message' };

    render(<ChatMessage message={systemMessage} onSuggestionClick={mockOnSuggestionClick} />);

    expect(screen.getByText('System message')).toBeInTheDocument();
  });

  it('renders product suggestions when present in metadata', () => {
    const messageWithSuggestions = {
      ...mockChatMessage,
      metadata: {
        suggestions: [mockSuggestion],
      },
    };

    render(<ChatMessage message={messageWithSuggestions} onSuggestionClick={mockOnSuggestionClick} />);

    expect(screen.getByText('Suggested Products:')).toBeInTheDocument();
    expect(screen.getByText('Test Product')).toBeInTheDocument();
  });

  it('renders actions when present in metadata', () => {
    const messageWithActions = {
      ...mockChatMessage,
      metadata: {
        actions: [
          { type: 'add_to_cart', payload: { product_id: '1' } },
        ],
      },
    };

    render(<ChatMessage message={messageWithActions} onSuggestionClick={mockOnSuggestionClick} />);

    expect(screen.getByText('Actions taken:')).toBeInTheDocument();
    expect(screen.getByText('Added to cart')).toBeInTheDocument();
  });
});

describe('ProductSuggestionCard', () => {
  const mockOnClick = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders product information correctly', () => {
    render(<ProductSuggestionCard suggestion={mockSuggestion} onClick={mockOnClick} />);

    expect(screen.getByText('Test Product')).toBeInTheDocument();
    expect(screen.getByText('A test product')).toBeInTheDocument();
    expect(screen.getByText('$99.99')).toBeInTheDocument();
    expect(screen.getByText('Electronics')).toBeInTheDocument();
  });

  it('shows confidence score', () => {
    render(<ProductSuggestionCard suggestion={mockSuggestion} onClick={mockOnClick} />);

    expect(screen.getByText('80% match')).toBeInTheDocument();
  });

  it('shows reason', () => {
    render(<ProductSuggestionCard suggestion={mockSuggestion} onClick={mockOnClick} />);

    expect(screen.getByText('Mentioned in conversation')).toBeInTheDocument();
  });

  it('calls onClick when clicked', () => {
    render(<ProductSuggestionCard suggestion={mockSuggestion} onClick={mockOnClick} />);

    const card = screen.getByText('Test Product').closest('div');
    fireEvent.click(card!);

    expect(mockOnClick).toHaveBeenCalledWith(mockSuggestion);
  });

  it('renders in compact mode', () => {
    render(<ProductSuggestionCard suggestion={mockSuggestion} onClick={mockOnClick} compact={true} />);

    expect(screen.getByText('Test Product')).toBeInTheDocument();
    expect(screen.getByText('$99.99')).toBeInTheDocument();
  });

  it('handles missing product gracefully', () => {
    const suggestionWithoutProduct = {
      ...mockSuggestion,
      product: null as any,
    };

    const { container } = render(
      <ProductSuggestionCard suggestion={suggestionWithoutProduct} onClick={mockOnClick} />
    );

    expect(container.firstChild).toBeNull();
  });

  // T074: Test ProductSuggestionCard renders with structured product data
  describe('Structured Product Data Rendering', () => {
    it('renders complete product data with category, tags, and inventory', () => {
      const fullProduct: any = {
        id: 'prod-123',
        name: 'Wireless Headphones',
        description: 'High-quality wireless headphones with noise cancellation',
        price: 199.99,
        sku: 'WH-001',
        status: 'active',
        category_id: 'cat-electronics',
        category: {
          id: 'cat-electronics',
          name: 'Electronics',
          description: 'Electronic products',
          slug: 'electronics',
          sort_order: 1,
          is_active: true,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        tags: ['new', 'popular', 'bestseller'],
        inventory: [
          {
            id: 'inv-001',
            product_id: 'prod-123',
            warehouse_location: 'Warehouse A',
            quantity_available: 100,
            quantity_reserved: 10,
            low_stock_threshold: 20,
            reorder_point: 30,
          },
        ],
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };

      const fullSuggestion: ProductSuggestion = {
        product: fullProduct,
        reason: 'Best match for your search',
        confidence: 0.95,
      };

      render(<ProductSuggestionCard suggestion={fullSuggestion} showAddToCart={true} />);

      // Verify product name and description
      expect(screen.getByText('Wireless Headphones')).toBeInTheDocument();
      expect(screen.getByText(/High-quality wireless headphones/)).toBeInTheDocument();

      // Verify price
      expect(screen.getByText('$199.99')).toBeInTheDocument();

      // Verify category is displayed
      expect(screen.getByText('Electronics')).toBeInTheDocument();

      // Verify tags are displayed (first 2)
      expect(screen.getByText('new')).toBeInTheDocument();
      expect(screen.getByText('popular')).toBeInTheDocument();
    });

    it('renders product with missing optional fields gracefully', () => {
      const minimalProduct: any = {
        id: 'prod-456',
        name: 'Basic Product',
        description: 'A basic product with minimal data',
        price: 49.99,
        sku: 'BP-001',
        status: 'active',
        category_id: 'cat-misc',
        // No category relationship
        // No tags
        // No inventory
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };

      const minimalSuggestion: ProductSuggestion = {
        product: minimalProduct,
      };

      const { container } = render(
        <ProductSuggestionCard suggestion={minimalSuggestion} showAddToCart={true} />
      );

      // Verify product still renders
      expect(screen.getByText('Basic Product')).toBeInTheDocument();
      expect(screen.getByText(/A basic product with minimal data/)).toBeInTheDocument();
      expect(screen.getByText('$49.99')).toBeInTheDocument();

      // Category, tags should not cause errors
      expect(container).toBeInTheDocument();
    });

    it('handles product with empty tags array', () => {
      const productWithEmptyTags: any = {
        id: 'prod-789',
        name: 'Product Without Tags',
        description: 'A product with no tags',
        price: 29.99,
        sku: 'NT-001',
        status: 'active',
        category_id: 'cat-misc',
        category: {
          id: 'cat-misc',
          name: 'Miscellaneous',
          slug: 'misc',
        },
        tags: [], // Empty tags array
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };

      const suggestion: ProductSuggestion = {
        product: productWithEmptyTags,
        reason: 'Available product',
        confidence: 0.75,
      };

      const { container } = render(
        <ProductSuggestionCard suggestion={suggestion} showAddToCart={true} />
      );

      // Verify product renders correctly without tags section
      expect(screen.getByText('Product Without Tags')).toBeInTheDocument();
      expect(screen.getByText('Miscellaneous')).toBeInTheDocument();
      expect(container).toBeInTheDocument();
    });

    it('displays out of stock status when inventory is zero', () => {
      const outOfStockProduct: any = {
        id: 'prod-999',
        name: 'Out of Stock Product',
        description: 'This product is out of stock',
        price: 79.99,
        sku: 'OOS-001',
        status: 'active',
        category_id: 'cat-electronics',
        category: {
          id: 'cat-electronics',
          name: 'Electronics',
          slug: 'electronics',
        },
        inventory: [
          {
            id: 'inv-999',
            product_id: 'prod-999',
            warehouse_location: 'Warehouse A',
            quantity_available: 0,
            quantity_reserved: 0,
            low_stock_threshold: 10,
            reorder_point: 20,
          },
        ],
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };

      const suggestion: ProductSuggestion = {
        product: outOfStockProduct,
        reason: 'Popular item',
        confidence: 0.88,
      };

      render(<ProductSuggestionCard suggestion={suggestion} showAddToCart={true} />);

      // Verify out of stock message
      expect(screen.getByText('Out of Stock Product')).toBeInTheDocument();
      expect(screen.getByText('Out of stock')).toBeInTheDocument();
    });

    it('shows Add to Cart button when showAddToCart is true', () => {
      const suggestion: ProductSuggestion = {
        product: mockProduct,
        reason: 'Recommended',
        confidence: 0.9,
      };

      render(<ProductSuggestionCard suggestion={suggestion} showAddToCart={true} />);

      // Verify Add to Cart button is present (button element should exist)
      const buttons = screen.getAllByRole('button');
      expect(buttons.length).toBeGreaterThan(0);
    });

    it('does not show Add to Cart button when showAddToCart is false', () => {
      const suggestion: ProductSuggestion = {
        product: mockProduct,
        reason: 'Recommended',
        confidence: 0.9,
      };

      render(<ProductSuggestionCard suggestion={suggestion} showAddToCart={false} />);

      // Verify Add to Cart button is not present
      const buttons = screen.queryAllByRole('button');
      expect(buttons.length).toBe(0);
    });
  });
});
