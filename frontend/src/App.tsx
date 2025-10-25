import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

// Layout components
import Layout from './components/Layout';
import Header from './components/Header';
import Footer from './components/Footer';

// Page components
import HomePage from './pages/HomePage';
import ProductsPage from './pages/ProductsPage';
import ProductDetailPage from './pages/ProductDetailPage';
import CartPage from './pages/CartPage';
import CheckoutPage from './pages/CheckoutPage';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import ProfilePage from './pages/ProfilePage';
import OrdersPage from './pages/OrdersPage';
import OrderDetailPage from './pages/OrderDetailPage';
import NotFoundPage from './pages/NotFoundPage';

// Context providers
import { AuthProvider } from './contexts/AuthContext';
import { CartProvider } from './contexts/CartContext';
import { NotificationProvider } from './contexts/NotificationContext';

// Create a client
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5 minutes
      gcTime: 10 * 60 * 1000, // 10 minutes
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

function App() {
  console.log('App component rendering...');
  
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <CartProvider>
          <NotificationProvider>
            <Router>
              <div className="min-h-screen bg-gray-50">
                <Header />
                <main className="flex-1">
                  <Routes>
                    {/* Public routes */}
                    <Route path="/" element={<HomePage />} />
                    <Route path="/products" element={<ProductsPage />} />
                    <Route path="/products/:id" element={<ProductDetailPage />} />
                    <Route path="/categories/:slug" element={<ProductsPage />} />
                    <Route path="/search" element={<ProductsPage />} />
                    
                    {/* Auth routes */}
                    <Route path="/login" element={<LoginPage />} />
                    <Route path="/register" element={<RegisterPage />} />
                    
                    {/* Protected routes */}
                    <Route path="/cart" element={
                      <Layout>
                        <CartPage />
                      </Layout>
                    } />
                    <Route path="/checkout" element={
                      <Layout>
                        <CheckoutPage />
                      </Layout>
                    } />
                    <Route path="/profile" element={
                      <Layout>
                        <ProfilePage />
                      </Layout>
                    } />
                    <Route path="/orders" element={
                      <Layout>
                        <OrdersPage />
                      </Layout>
                    } />
                    <Route path="/orders/:id" element={
                      <Layout>
                        <OrderDetailPage />
                      </Layout>
                    } />
                    
                    {/* 404 route */}
                    <Route path="*" element={<NotFoundPage />} />
                  </Routes>
                </main>
                <Footer />
              </div>
            </Router>
          </NotificationProvider>
        </CartProvider>
      </AuthProvider>
    </QueryClientProvider>
  );
}

export default App;