import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { useCart } from '../context/CartContext';

export function Navbar() {
  const { user, logout, isAuthenticated } = useAuth();
  const { totalItems } = useCart();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <nav className="bg-white border-b border-gray-200 px-6 py-3 flex items-center justify-between shadow-sm">
      <Link to="/products" className="text-xl font-bold text-indigo-600 hover:text-indigo-700">
        ShopApp
      </Link>

      <div className="flex items-center gap-4">
        {isAuthenticated && (
          <>
            <Link to="/products" className="text-gray-600 hover:text-indigo-600 text-sm font-medium">
              Products
            </Link>
            <Link to="/orders" className="text-gray-600 hover:text-indigo-600 text-sm font-medium">
              Orders
            </Link>
            <Link to="/cart" className="relative text-gray-600 hover:text-indigo-600">
              <span className="text-xl">🛒</span>
              {totalItems > 0 && (
                <span className="absolute -top-2 -right-2 bg-indigo-600 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center font-bold">
                  {totalItems}
                </span>
              )}
            </Link>
            {user && (
              <span className="text-sm text-gray-700 font-medium">{user.name}</span>
            )}
            <button
              onClick={handleLogout}
              className="text-sm text-red-500 hover:text-red-700 font-medium"
            >
              Logout
            </button>
          </>
        )}
        {!isAuthenticated && (
          <>
            <Link to="/login" className="text-sm text-indigo-600 hover:text-indigo-700 font-medium">
              Login
            </Link>
            <Link to="/register" className="text-sm bg-indigo-600 text-white px-3 py-1 rounded hover:bg-indigo-700">
              Register
            </Link>
          </>
        )}
      </div>
    </nav>
  );
}
