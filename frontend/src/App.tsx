import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { CartProvider } from './context/CartContext';
import { Navbar } from './components/Navbar';
import { ProtectedRoute } from './components/ProtectedRoute';
import { LoginPage } from './pages/LoginPage';
import { RegisterPage } from './pages/RegisterPage';
import { ProductsPage } from './pages/ProductsPage';
import { ProductDetailPage } from './pages/ProductDetailPage';
import { CartPage } from './pages/CartPage';
import { OrdersPage } from './pages/OrdersPage';
import { OrderDetailPage } from './pages/OrderDetailPage';

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <CartProvider>
          <div className="min-h-screen bg-gray-50">
            <Navbar />
            <Routes>
              <Route path="/" element={<Navigate to="/products" replace />} />
              <Route path="/login" element={<LoginPage />} />
              <Route path="/register" element={<RegisterPage />} />
              <Route path="/products" element={
                <ProtectedRoute><ProductsPage /></ProtectedRoute>
              } />
              <Route path="/products/:id" element={
                <ProtectedRoute><ProductDetailPage /></ProtectedRoute>
              } />
              <Route path="/cart" element={
                <ProtectedRoute><CartPage /></ProtectedRoute>
              } />
              <Route path="/orders" element={
                <ProtectedRoute><OrdersPage /></ProtectedRoute>
              } />
              <Route path="/orders/:id" element={
                <ProtectedRoute><OrderDetailPage /></ProtectedRoute>
              } />
            </Routes>
          </div>
        </CartProvider>
      </AuthProvider>
    </BrowserRouter>
  );
}

export default App;
