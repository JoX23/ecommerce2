export interface User {
  id: string;
  email: string;
  name: string;
  createdAt: string;
}

export interface Product {
  id: string;
  sku: string;
  name: string;
  price: number;
  stock: number;
  description?: string;
  imageUrl?: string;
  status: 'draft' | 'published' | 'archived';
  createdAt: string;
}

export interface OrderItem {
  productId: string;
  productName: string;
  qty: number;
  unitPrice: number;
  subtotal: number;
}

export interface Order {
  id: string;
  userId: string;
  status: 'pending' | 'confirmed' | 'shipped' | 'delivered' | 'cancelled';
  total: number;
  items: OrderItem[];
  createdAt: string;
}

export interface CartItem {
  product: Product;
  qty: number;
}
