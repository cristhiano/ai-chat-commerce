import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useCart } from '../../contexts/CartContext';
import { useAuth } from '../../contexts/AuthContext';
import { apiService } from '../../services/api';
import { formatCurrency } from '../../utils';

interface CheckoutFormProps {
  onOrderCreated?: (orderId: string) => void;
}

const CheckoutForm: React.FC<CheckoutFormProps> = ({ onOrderCreated }) => {
  const navigate = useNavigate();
  const { cart, clearCart } = useCart();
  const { user, isAuthenticated } = useAuth();
  
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});
  
  // Form state
  const [formData, setFormData] = useState({
    // Shipping Address
    shippingFirstName: '',
    shippingLastName: '',
    shippingAddress1: '',
    shippingAddress2: '',
    shippingCity: '',
    shippingState: '',
    shippingZip: '',
    shippingCountry: 'US',
    
    // Billing Address
    billingSameAsShipping: true,
    billingFirstName: '',
    billingLastName: '',
    billingAddress1: '',
    billingAddress2: '',
    billingCity: '',
    billingState: '',
    billingZip: '',
    billingCountry: 'US',
    
    // Payment
    paymentMethod: 'card',
    cardNumber: '',
    expiryMonth: '',
    expiryYear: '',
    cvv: '',
    cardholderName: '',
    
    // Contact
    email: user?.email || '',
    phone: '',
    
    // Order
    notes: ''
  });

  // Handle input changes
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    const { name, value, type } = e.target;
    
    if (type === 'checkbox') {
      const checked = (e.target as HTMLInputElement).checked;
      setFormData(prev => ({
        ...prev,
        [name]: checked
      }));
      
      // If billing same as shipping is checked, copy shipping data
      if (name === 'billingSameAsShipping' && checked) {
        setFormData(prev => ({
          ...prev,
          billingFirstName: prev.shippingFirstName,
          billingLastName: prev.shippingLastName,
          billingAddress1: prev.shippingAddress1,
          billingAddress2: prev.shippingAddress2,
          billingCity: prev.shippingCity,
          billingState: prev.shippingState,
          billingZip: prev.shippingZip,
          billingCountry: prev.shippingCountry,
        }));
      }
    } else {
      setFormData(prev => ({
        ...prev,
        [name]: value
      }));
    }
    
    // Clear error when user starts typing
    if (errors[name]) {
      setErrors(prev => ({
        ...prev,
        [name]: ''
      }));
    }
  };

  // Validate form
  const validateForm = () => {
    const newErrors: Record<string, string> = {};
    
    // Required fields
    const requiredFields = [
      'shippingFirstName', 'shippingLastName', 'shippingAddress1',
      'shippingCity', 'shippingState', 'shippingZip', 'email'
    ];
    
    requiredFields.forEach(field => {
      if (!formData[field as keyof typeof formData]) {
        newErrors[field] = 'This field is required';
      }
    });
    
    // Email validation
    if (formData.email && !/\S+@\S+\.\S+/.test(formData.email)) {
      newErrors.email = 'Please enter a valid email address';
    }
    
    // Billing address validation if different from shipping
    if (!formData.billingSameAsShipping) {
      const billingRequiredFields = [
        'billingFirstName', 'billingLastName', 'billingAddress1',
        'billingCity', 'billingState', 'billingZip'
      ];
      
      billingRequiredFields.forEach(field => {
        if (!formData[field as keyof typeof formData]) {
          newErrors[field] = 'This field is required';
        }
      });
    }
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // Handle form submission
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!cart || cart.items.length === 0) {
      alert('Your cart is empty');
      return;
    }
    
    if (!validateForm()) {
      return;
    }
    
    setIsSubmitting(true);
    
    try {
      // Prepare order data
      const orderData = {
        user_id: user?.id || undefined,
        session_id: localStorage.getItem('session_id') || undefined,
        items: cart.items.map(item => ({
          product_id: item.product_id,
          variant_id: item.variant_id,
          quantity: item.quantity
        })),
        shipping_address: {
          first_name: formData.shippingFirstName,
          last_name: formData.shippingLastName,
          address1: formData.shippingAddress1,
          address2: formData.shippingAddress2,
          city: formData.shippingCity,
          state: formData.shippingState,
          zip: formData.shippingZip,
          country: formData.shippingCountry,
        },
        billing_address: formData.billingSameAsShipping ? {
          first_name: formData.shippingFirstName,
          last_name: formData.shippingLastName,
          address1: formData.shippingAddress1,
          address2: formData.shippingAddress2,
          city: formData.shippingCity,
          state: formData.shippingState,
          zip: formData.shippingZip,
          country: formData.shippingCountry,
        } : {
          first_name: formData.billingFirstName,
          last_name: formData.billingLastName,
          address1: formData.billingAddress1,
          address2: formData.billingAddress2,
          city: formData.billingCity,
          state: formData.billingState,
          zip: formData.billingZip,
          country: formData.billingCountry,
        },
        payment_method: formData.paymentMethod,
        notes: formData.notes
      };

      // Create order
      const orderResponse = await apiService.createOrder(orderData);
      
      if (orderResponse.data) {
        // Create payment intent
        const paymentResponse = await apiService.createPaymentIntent({
          order_id: orderResponse.data.id,
          amount: Math.round(cart.total_amount * 100), // Convert to cents
          currency: 'USD',
          description: `Order ${orderResponse.data.order_number}`,
          metadata: {
            order_number: orderResponse.data.order_number
          }
        });

        if (paymentResponse.data) {
          // In a real implementation, you would integrate with Stripe Elements
          // For now, we'll simulate a successful payment
          await apiService.confirmPayment({
            payment_intent_id: paymentResponse.data.id,
            order_id: orderResponse.data.id
          });

          // Clear cart
          await clearCart();
          
          // Redirect to order confirmation
          onOrderCreated?.(orderResponse.data.id);
          navigate(`/orders/${orderResponse.data.id}`);
        }
      }
    } catch (error) {
      console.error('Checkout failed:', error);
      alert('Checkout failed. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!cart || cart.items.length === 0) {
    return (
      <div className="text-center py-12">
        <div className="text-gray-600 text-lg mb-4">Your cart is empty</div>
        <button
          onClick={() => navigate('/products')}
          className="bg-blue-600 text-white px-6 py-3 rounded-lg hover:bg-blue-700 transition-colors"
        >
          Continue Shopping
        </button>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <h1 className="text-3xl font-bold text-gray-900 mb-8">Checkout</h1>

      <form onSubmit={handleSubmit} className="space-y-8">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Left Column - Form */}
          <div className="space-y-8">
            {/* Contact Information */}
            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h2 className="text-xl font-semibold text-gray-900 mb-4">
                Contact Information
              </h2>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Email Address *
                  </label>
                  <input
                    type="email"
                    name="email"
                    value={formData.email}
                    onChange={handleInputChange}
                    className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                      errors.email ? 'border-red-500' : 'border-gray-300'
                    }`}
                    required
                  />
                  {errors.email && (
                    <p className="text-red-500 text-sm mt-1">{errors.email}</p>
                  )}
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Phone Number
                  </label>
                  <input
                    type="tel"
                    name="phone"
                    value={formData.phone}
                    onChange={handleInputChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>
            </div>

            {/* Shipping Address */}
            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h2 className="text-xl font-semibold text-gray-900 mb-4">
                Shipping Address
              </h2>
              
              <div className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      First Name *
                    </label>
                    <input
                      type="text"
                      name="shippingFirstName"
                      value={formData.shippingFirstName}
                      onChange={handleInputChange}
                      className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                        errors.shippingFirstName ? 'border-red-500' : 'border-gray-300'
                      }`}
                      required
                    />
                    {errors.shippingFirstName && (
                      <p className="text-red-500 text-sm mt-1">{errors.shippingFirstName}</p>
                    )}
                  </div>
                  
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Last Name *
                    </label>
                    <input
                      type="text"
                      name="shippingLastName"
                      value={formData.shippingLastName}
                      onChange={handleInputChange}
                      className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                        errors.shippingLastName ? 'border-red-500' : 'border-gray-300'
                      }`}
                      required
                    />
                    {errors.shippingLastName && (
                      <p className="text-red-500 text-sm mt-1">{errors.shippingLastName}</p>
                    )}
                  </div>
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Address Line 1 *
                  </label>
                  <input
                    type="text"
                    name="shippingAddress1"
                    value={formData.shippingAddress1}
                    onChange={handleInputChange}
                    className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                      errors.shippingAddress1 ? 'border-red-500' : 'border-gray-300'
                    }`}
                    required
                  />
                  {errors.shippingAddress1 && (
                    <p className="text-red-500 text-sm mt-1">{errors.shippingAddress1}</p>
                  )}
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Address Line 2
                  </label>
                  <input
                    type="text"
                    name="shippingAddress2"
                    value={formData.shippingAddress2}
                    onChange={handleInputChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      City *
                    </label>
                    <input
                      type="text"
                      name="shippingCity"
                      value={formData.shippingCity}
                      onChange={handleInputChange}
                      className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                        errors.shippingCity ? 'border-red-500' : 'border-gray-300'
                      }`}
                      required
                    />
                    {errors.shippingCity && (
                      <p className="text-red-500 text-sm mt-1">{errors.shippingCity}</p>
                    )}
                  </div>
                  
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      State *
                    </label>
                    <input
                      type="text"
                      name="shippingState"
                      value={formData.shippingState}
                      onChange={handleInputChange}
                      className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                        errors.shippingState ? 'border-red-500' : 'border-gray-300'
                      }`}
                      required
                    />
                    {errors.shippingState && (
                      <p className="text-red-500 text-sm mt-1">{errors.shippingState}</p>
                    )}
                  </div>
                  
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      ZIP Code *
                    </label>
                    <input
                      type="text"
                      name="shippingZip"
                      value={formData.shippingZip}
                      onChange={handleInputChange}
                      className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                        errors.shippingZip ? 'border-red-500' : 'border-gray-300'
                      }`}
                      required
                    />
                    {errors.shippingZip && (
                      <p className="text-red-500 text-sm mt-1">{errors.shippingZip}</p>
                    )}
                  </div>
                </div>
              </div>
            </div>

            {/* Billing Address */}
            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h2 className="text-xl font-semibold text-gray-900 mb-4">
                Billing Address
              </h2>
              
              <div className="mb-4">
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    name="billingSameAsShipping"
                    checked={formData.billingSameAsShipping}
                    onChange={handleInputChange}
                    className="mr-2"
                  />
                  <span className="text-sm text-gray-700">
                    Same as shipping address
                  </span>
                </label>
              </div>
              
              {!formData.billingSameAsShipping && (
                <div className="space-y-4">
                  {/* Billing address fields - similar to shipping */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        First Name *
                      </label>
                      <input
                        type="text"
                        name="billingFirstName"
                        value={formData.billingFirstName}
                        onChange={handleInputChange}
                        className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                          errors.billingFirstName ? 'border-red-500' : 'border-gray-300'
                        }`}
                        required={!formData.billingSameAsShipping}
                      />
                      {errors.billingFirstName && (
                        <p className="text-red-500 text-sm mt-1">{errors.billingFirstName}</p>
                      )}
                    </div>
                    
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Last Name *
                      </label>
                      <input
                        type="text"
                        name="billingLastName"
                        value={formData.billingLastName}
                        onChange={handleInputChange}
                        className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                          errors.billingLastName ? 'border-red-500' : 'border-gray-300'
                        }`}
                        required={!formData.billingSameAsShipping}
                      />
                      {errors.billingLastName && (
                        <p className="text-red-500 text-sm mt-1">{errors.billingLastName}</p>
                      )}
                    </div>
                  </div>
                  
                  {/* Additional billing fields... */}
                </div>
              )}
            </div>

            {/* Order Notes */}
            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h2 className="text-xl font-semibold text-gray-900 mb-4">
                Order Notes
              </h2>
              
              <textarea
                name="notes"
                value={formData.notes}
                onChange={handleInputChange}
                rows={3}
                placeholder="Any special instructions for your order..."
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </div>

          {/* Right Column - Order Summary */}
          <div className="space-y-6">
            {/* Order Summary */}
            <div className="bg-white rounded-lg shadow-sm border sticky top-8">
              <div className="px-6 py-4 border-b border-gray-200">
                <h2 className="text-lg font-semibold text-gray-900">
                  Order Summary
                </h2>
              </div>
              
              <div className="px-6 py-4 space-y-4">
                {/* Cart Items */}
                <div className="space-y-3">
                  {cart.items.map((item) => (
                    <div key={`${item.product_id}-${item.variant_id || 'default'}`} className="flex justify-between">
                      <div className="flex-1">
                        <p className="text-sm font-medium text-gray-900">{item.product_name}</p>
                        <p className="text-xs text-gray-500">Qty: {item.quantity}</p>
                      </div>
                      <p className="text-sm font-medium text-gray-900">
                        {formatCurrency(item.total_price)}
                      </p>
                    </div>
                  ))}
                </div>

                {/* Totals */}
                <div className="border-t border-gray-200 pt-4 space-y-2">
                  <div className="flex justify-between">
                    <span className="text-gray-600">Subtotal</span>
                    <span className="font-medium">{formatCurrency(cart.subtotal)}</span>
                  </div>
                  
                  <div className="flex justify-between">
                    <span className="text-gray-600">Tax</span>
                    <span className="font-medium">{formatCurrency(cart.tax_amount)}</span>
                  </div>
                  
                  <div className="flex justify-between">
                    <span className="text-gray-600">Shipping</span>
                    <span className="font-medium">{formatCurrency(cart.shipping_amount)}</span>
                  </div>
                  
                  <div className="border-t border-gray-200 pt-2">
                    <div className="flex justify-between">
                      <span className="text-lg font-semibold text-gray-900">Total</span>
                      <span className="text-lg font-semibold text-blue-600">
                        {formatCurrency(cart.total_amount)}
                      </span>
                    </div>
                  </div>
                </div>

                {/* Submit Button */}
                <button
                  type="submit"
                  disabled={isSubmitting}
                  className="w-full bg-blue-600 text-white py-3 px-6 rounded-lg font-semibold hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {isSubmitting ? 'Processing...' : 'Complete Order'}
                </button>
              </div>
            </div>
          </div>
        </div>
      </form>
    </div>
  );
};

export default CheckoutForm;
