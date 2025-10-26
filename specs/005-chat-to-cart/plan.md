# Project Plan: Structured Product Data in Chat Responses

## Constitution Check

This plan MUST align with the following constitutional principles:
- **Code Quality:** All proposed features will maintain high standards of readability and maintainability by creating clear, well-structured data contracts
- **Testing Standards:** Comprehensive testing strategy will include unit tests for data serialization/deserialization and integration tests for WebSocket message handling
- **User Experience Consistency:** Design decisions will follow established product card patterns from ProductsPage, ensuring visual consistency across the application
- **Performance Requirements:** Chat responses with product data will be efficiently structured to minimize payload size while maintaining full product information

### Gate Evaluation
- ✅ **Code Quality Gate:** Will maintain existing patterns and create clear data structure contracts
- ✅ **Testing Gate:** Will achieve 80%+ code coverage with unit and integration tests for message handling
- ✅ **UX Consistency Gate:** Cards will match existing product card design system
- ✅ **Performance Gate:** Structured data payloads will be optimized for minimal size and fast parsing

## Project Overview

**Project Name:** Structured Product Data in Chat Responses
**Version:** 1.0
**Start Date:** 2024-12-26
**Target Completion:** 2024-12-27
**Project Manager:** AI Assistant

## Objectives

### Primary Goals
- Separate product suggestion data from text message content in chat responses
- Send product information as structured JSON via WebSocket metadata
- Ensure frontend can parse and render product cards from structured data
- Maintain backward compatibility with existing chat message format

### Success Metrics
- Product suggestions appear as visual cards, not plain text
- Chat messages with products include structured data in metadata
- Frontend successfully parses and renders all product fields
- No breaking changes to existing chat functionality

## Scope

### In Scope
- Modify backend ChatResponse to include ProductSuggestion objects
- Update WebSocket message structure to include product metadata
- Ensure frontend ProductSuggestionCard receives all required fields
- Add proper TypeScript interfaces for typed product data
- Handle edge cases (missing fields, partial data)

### Out of Scope
- Product recommendation algorithm changes
- Cart functionality (already implemented)
- UI component redesign (already implemented)
- Chat history migration

## Technical Requirements

### Code Quality Standards
- Follow established Go coding patterns for struct definitions
- Use consistent JSON serialization approach
- Document all WebSocket message types
- Implement proper error handling for malformed data

### Testing Strategy
- Unit tests: JSON serialization/deserialization of ProductSuggestion
- Integration tests: WebSocket message delivery with product data
- E2E tests: Frontend renders product cards from structured data
- Test edge cases: missing fields, partial product data

### User Experience Requirements
- No visual disruption during transition to structured data
- Product cards render immediately when data arrives
- Graceful degradation if product data is incomplete
- Consistent product card appearance across all contexts

### Performance Targets
- Product data serialization completes within 10ms
- WebSocket message size kept under 50KB for product lists
- Frontend parsing completes within 50ms
- No UI blocking during data processing

## Data Structure Requirements

### Current State (Problem)
Currently, products are being included in the text message content itself:
```
Message: "I recommend Product X: $29.99, SKU: ABC-123..."
```

This makes it impossible for the frontend to extract structured product data.

### Target State (Solution)
Products should be sent as separate, structured data:

**WebSocket Message Structure:**
```json
{
  "type": "message",
  "data": {
    "id": "msg-123",
    "role": "assistant",
    "content": "Here are some product recommendations:",
    "metadata": {
      "suggestions": [
        {
          "product": {
            "id": "uuid",
            "name": "Product Name",
            "description": "Product description...",
            "price": 29.99,
            "category": {
              "id": "uuid",
              "name": "Electronics"
            },
            "tags": ["popular", "bestseller"],
            "inventory": [
              {
                "id": "uuid",
                "quantity_available": 10
              }
            ]
          },
          "reason": "Matches your search",
          "confidence": 0.85
        }
      ]
    }
  }
}
```

### Implementation Details

#### Backend Changes Required

1. **Chat Handler** (`backend/internal/handlers/chat_handler.go`)
   - Ensure suggestions are included in message metadata
   - Send suggestions as separate WebSocket message with type "suggestions"
   - Verify ProductSuggestion struct includes full Product data with relationships

2. **Chat Service** (`backend/internal/services/chat_service.go`)
   - Already generates ProductSuggestion objects ✓
   - Already includes them in ChatResponse ✓
   - Need to ensure category is preloaded for all products
   - Verify all product fields are populated

3. **Product Model** (`backend/internal/models/models.go`)
   - Already has relationships defined (Category, Inventory) ✓
   - Need to ensure Preload is called when fetching products for chat

#### Frontend Changes Required

1. **ChatInterface** (`frontend/src/components/chat/ChatInterface.tsx`)
   - Already handles WebSocket "suggestions" message type ✓
   - Already passes suggestions to state ✓
   - No changes needed

2. **ChatMessage** (`frontend/src/components/chat/ChatMessage.tsx`)
   - Already renders ProductSuggestionCard for suggestions in metadata ✓
   - Need to ensure it uses full card layout, not compact

3. **Type Definitions** (`frontend/src/types/index.ts`)
   - Verify ProductSuggestion interface matches backend
   - Ensure all fields are properly typed

## Timeline

### Phase 1: Backend Data Structure (Duration: 4 hours)
- **Deliverables:**
  - Verify ProductSuggestion includes full Product data
  - Ensure category relationship is preloaded
  - Test JSON serialization of product suggestions
- **Testing Requirements:**
  - Unit test: ProductSuggestion serialization
  - Integration test: WebSocket sends complete product data
  - Verify all product fields are included in response

### Phase 2: Frontend Data Consumption (Duration: 2 hours)
- **Deliverables:**
  - Verify ChatMessage renders full product cards
  - Test that all product fields display correctly
  - Ensure quantity editor and cart functionality work
- **Testing Requirements:**
  - E2E test: Product cards render with all data
  - Verify cart button functionality
  - Test mobile responsiveness

### Phase 3: Testing & Validation (Duration: 2 hours)
- **Deliverables:**
  - Full integration test suite passes
  - Verify no breaking changes
  - Validate performance metrics
- **Testing Requirements:**
  - All existing tests pass
  - New tests for structured data handling
  - Performance benchmarks met

## Risk Assessment

### Technical Risks
- **Risk 1**: Missing product fields cause frontend errors
  - **Probability**: Medium
  - **Impact**: Low - frontend has default handling
  - **Mitigation**: Comprehensive backend validation, frontend fallbacks

- **Risk 2**: Large product lists cause message size issues
  - **Probability**: Low
  - **Impact**: Low - limit to 5-10 products
  - **Mitigation**: Limit suggestions returned, compress if needed

### Performance Risks
- **Risk 1**: Serialization overhead in high-traffic scenarios
  - **Probability**: Low
  - **Impact**: Very Low - minimal serialization overhead
  - **Mitigation**: Efficient JSON library, caching

## Quality Assurance

### Code Review Process
- Backend struct definitions reviewed for completeness
- Frontend type safety verified
- JSON schema validation

### Testing Process
- All WebSocket message types tested
- Frontend rendering tested with various product data scenarios
- Performance testing for message size and parsing speed

### Compliance Monitoring
- Verify structured data meets performance targets
- Ensure no breaking changes to existing functionality
- Accessibility verified for product cards

## Resources

### Team Members
- Developer: AI Assistant

### Tools and Technologies
- Go: Backend serialization
- TypeScript: Frontend type definitions
- React: Component rendering
- WebSocket: Real-time communication

## Approval

**Project Sponsor:** TBD - 2024-12-26  
**Technical Lead:** Pending - 2024-12-26  
**Quality Assurance:** Pending - 2024-12-26
