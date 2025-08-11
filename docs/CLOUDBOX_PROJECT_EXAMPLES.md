# CloudBox Project Examples - BaaS voor Verschillende Projecttypes

## 1. E-commerce Platform

### Setup
```javascript
import { CloudBoxUniversalClient } from '@cloudbox/universal-sdk';

class EcommerceAPI extends CloudBoxUniversalClient {
  // Products
  products = {
    getAll: (filters = {}) => {
      return this.database.query('products', {
        status: 'active',
        ...filters
      }, {
        orderBy: { created_at: 'desc' }
      });
    },

    getByCategory: (categoryId) => {
      return this.database.query('products', {
        category_id: categoryId,
        status: 'active'
      });
    },

    create: (product) => this.database.create('products', {
      name: product.name,
      description: product.description,
      price: product.price,
      category_id: product.categoryId,
      sku: product.sku,
      stock: product.stock,
      images: JSON.stringify(product.images || []),
      status: 'active'
    }),

    updateStock: (id, stock) => this.database.update('products', id, { stock }),
    
    search: (query) => {
      return this.database.query('products', {
        $or: [
          { name: { $contains: query } },
          { description: { $contains: query } }
        ]
      });
    }
  };

  // Orders
  orders = {
    create: (order) => this.database.create('orders', {
      user_id: order.userId,
      items: JSON.stringify(order.items),
      subtotal: order.subtotal,
      tax: order.tax,
      shipping: order.shipping,
      total: order.total,
      status: 'pending',
      shipping_address: JSON.stringify(order.shippingAddress),
      billing_address: JSON.stringify(order.billingAddress)
    }),

    getByUser: (userId) => this.database.query('orders', { user_id: userId }),

    updateStatus: (id, status) => this.database.update('orders', id, { 
      status,
      updated_at: new Date().toISOString()
    }),

    getByStatus: (status) => this.database.query('orders', { status })
  };

  // Shopping Cart (session-based)
  cart = {
    add: (userId, productId, quantity) => this.database.create('cart_items', {
      user_id: userId,
      product_id: productId,
      quantity: quantity
    }),

    getByUser: (userId) => this.database.query('cart_items', { user_id: userId }),

    updateQuantity: (id, quantity) => this.database.update('cart_items', id, { quantity }),

    remove: (id) => this.database.delete('cart_items', id),

    clear: (userId) => {
      // Delete all cart items for user
      return this.functions.call('clear-cart', { user_id: userId });
    }
  };

  // Categories
  categories = {
    getAll: () => this.database.findMany('categories', {
      orderBy: { name: 'asc' }
    }),

    getHierarchy: () => this.functions.call('get-category-hierarchy'),

    create: (category) => this.database.create('categories', {
      name: category.name,
      description: category.description,
      parent_id: category.parentId || null,
      image_url: category.imageUrl || null
    })
  };
}

// Usage
const ecommerce = new EcommerceAPI({
  endpoint: 'https://your-cloudbox.com',
  apiKey: 'your-api-key',
  projectId: 'ecommerce-project'
});

// Get products in electronics category
const products = await ecommerce.products.getByCategory('electronics');

// Create an order
const order = await ecommerce.orders.create({
  userId: 'user123',
  items: [{ productId: 'prod1', quantity: 2, price: 29.99 }],
  total: 59.98
});
```

## 2. Blog Platform

### Setup
```javascript
class BlogAPI extends CloudBoxUniversalClient {
  // Posts
  posts = {
    getPublished: (limit = 10, offset = 0) => {
      return this.database.query('posts', {
        status: 'published'
      }, {
        orderBy: { published_at: 'desc' },
        limit,
        offset
      });
    },

    getBySlug: (slug) => {
      return this.database.query('posts', { slug }).then(posts => posts[0]);
    },

    getByAuthor: (authorId) => {
      return this.database.query('posts', { author_id: authorId });
    },

    getByTag: (tag) => {
      return this.database.query('posts', {
        tags: { $contains: tag }
      });
    },

    create: (post) => this.database.create('posts', {
      title: post.title,
      slug: this.utils.generateSlug(post.title),
      content: post.content,
      excerpt: post.excerpt,
      author_id: post.authorId,
      tags: JSON.stringify(post.tags || []),
      featured_image: post.featuredImage || null,
      status: 'draft'
    }),

    publish: (id) => this.database.update('posts', id, {
      status: 'published',
      published_at: new Date().toISOString()
    }),

    incrementViews: (id) => this.functions.call('increment-post-views', { post_id: id })
  };

  // Comments
  comments = {
    getByPost: (postId) => {
      return this.database.query('comments', {
        post_id: postId,
        status: 'approved'
      }, {
        orderBy: { created_at: 'desc' }
      });
    },

    create: (comment) => this.database.create('comments', {
      post_id: comment.postId,
      author_name: comment.authorName,
      author_email: comment.authorEmail,
      content: comment.content,
      status: 'pending' // Moderation required
    }),

    approve: (id) => this.database.update('comments', id, { status: 'approved' }),

    reject: (id) => this.database.update('comments', id, { status: 'rejected' })
  };

  // Authors
  authors = {
    getAll: () => this.database.findMany('authors'),

    getBySlug: (slug) => {
      return this.database.query('authors', { slug }).then(authors => authors[0]);
    },

    create: (author) => this.database.create('authors', {
      name: author.name,
      slug: this.utils.generateSlug(author.name),
      bio: author.bio,
      avatar: author.avatar,
      social_links: JSON.stringify(author.socialLinks || {})
    })
  };

  // Newsletter
  newsletter = {
    subscribe: (email) => this.database.create('subscribers', {
      email,
      status: 'active',
      subscribed_at: new Date().toISOString()
    }),

    unsubscribe: (email) => this.database.query('subscribers', { email })
      .then(subs => subs[0] && this.database.update('subscribers', subs[0].id, { status: 'unsubscribed' }))
  };
}

// Usage
const blog = new BlogAPI({
  endpoint: 'https://your-cloudbox.com',
  apiKey: 'your-api-key',
  projectId: 'blog-project'
});

// Get recent posts
const posts = await blog.posts.getPublished(5);

// Create a new post
const newPost = await blog.posts.create({
  title: 'Getting Started with CloudBox',
  content: 'CloudBox is a powerful BaaS...',
  excerpt: 'Learn how to use CloudBox...',
  authorId: 'author1',
  tags: ['cloudbox', 'tutorial']
});
```

## 3. SaaS Application

### Setup
```javascript
class SaaSAPI extends CloudBoxUniversalClient {
  // Organizations
  organizations = {
    create: (org) => this.database.create('organizations', {
      name: org.name,
      slug: this.utils.generateSlug(org.name),
      plan: org.plan || 'free',
      max_users: org.maxUsers || 5,
      settings: JSON.stringify(org.settings || {})
    }),

    getByUser: (userId) => {
      return this.database.query('user_organizations', {
        user_id: userId
      }).then(userOrgs => {
        const orgIds = userOrgs.map(uo => uo.data.organization_id);
        return this.database.query('organizations', {
          id: { $in: orgIds }
        });
      });
    },

    addUser: (orgId, userId, role = 'member') => {
      return this.database.create('user_organizations', {
        organization_id: orgId,
        user_id: userId,
        role: role
      });
    },

    updatePlan: (id, plan) => this.database.update('organizations', id, { plan })
  };

  // Projects (within organizations)
  projects = {
    create: (project) => this.database.create('projects', {
      name: project.name,
      description: project.description,
      organization_id: project.organizationId,
      status: 'active'
    }),

    getByOrganization: (orgId) => {
      return this.database.query('projects', {
        organization_id: orgId,
        status: 'active'
      });
    }
  };

  // Tasks
  tasks = {
    create: (task) => this.database.create('tasks', {
      title: task.title,
      description: task.description,
      project_id: task.projectId,
      assigned_to: task.assignedTo,
      status: 'todo',
      priority: task.priority || 'medium',
      due_date: task.dueDate || null
    }),

    getByProject: (projectId) => {
      return this.database.query('tasks', {
        project_id: projectId
      }, {
        orderBy: { created_at: 'desc' }
      });
    },

    updateStatus: (id, status) => this.database.update('tasks', id, { status }),

    getByUser: (userId) => {
      return this.database.query('tasks', {
        assigned_to: userId
      });
    }
  };

  // Billing
  billing = {
    createSubscription: (subscription) => this.database.create('subscriptions', {
      organization_id: subscription.organizationId,
      plan: subscription.plan,
      status: 'active',
      current_period_start: subscription.currentPeriodStart,
      current_period_end: subscription.currentPeriodEnd,
      stripe_subscription_id: subscription.stripeSubscriptionId
    }),

    getByOrganization: (orgId) => {
      return this.database.query('subscriptions', {
        organization_id: orgId,
        status: 'active'
      }).then(subs => subs[0]);
    },

    createInvoice: (invoice) => this.database.create('invoices', invoice)
  };
}

// Usage
const saas = new SaaSAPI({
  endpoint: 'https://your-cloudbox.com',
  apiKey: 'your-api-key',
  projectId: 'saas-project'
});

// Create organization
const org = await saas.organizations.create({
  name: 'Acme Corp',
  plan: 'pro'
});

// Create project
const project = await saas.projects.create({
  name: 'Website Redesign',
  organizationId: org.id
});
```

## 4. Social Media App

### Setup
```javascript
class SocialAPI extends CloudBoxUniversalClient {
  // Posts
  posts = {
    getFeed: (userId, limit = 20) => {
      // Get posts from followed users
      return this.functions.call('get-user-feed', {
        user_id: userId,
        limit
      });
    },

    create: (post) => this.database.create('posts', {
      user_id: post.userId,
      content: post.content,
      images: JSON.stringify(post.images || []),
      type: post.type || 'text', // text, image, video
      visibility: post.visibility || 'public'
    }),

    getByUser: (userId) => {
      return this.database.query('posts', {
        user_id: userId,
        visibility: 'public'
      }, {
        orderBy: { created_at: 'desc' }
      });
    }
  };

  // Likes
  likes = {
    like: (userId, postId) => this.database.create('likes', {
      user_id: userId,
      post_id: postId
    }),

    unlike: (userId, postId) => {
      return this.database.query('likes', {
        user_id: userId,
        post_id: postId
      }).then(likes => {
        return likes.length > 0 ? this.database.delete('likes', likes[0].id) : null;
      });
    },

    getCount: (postId) => {
      return this.database.count('likes', { post_id: postId });
    }
  };

  // Follows
  follows = {
    follow: (followerId, followingId) => this.database.create('follows', {
      follower_id: followerId,
      following_id: followingId
    }),

    unfollow: (followerId, followingId) => {
      return this.database.query('follows', {
        follower_id: followerId,
        following_id: followingId
      }).then(follows => {
        return follows.length > 0 ? this.database.delete('follows', follows[0].id) : null;
      });
    },

    getFollowers: (userId) => {
      return this.database.query('follows', {
        following_id: userId
      });
    },

    getFollowing: (userId) => {
      return this.database.query('follows', {
        follower_id: userId
      });
    }
  };
}
```

## 5. Event Management System

### Setup
```javascript
class EventAPI extends CloudBoxUniversalClient {
  events = {
    create: (event) => this.database.create('events', {
      title: event.title,
      description: event.description,
      start_date: event.startDate,
      end_date: event.endDate,
      location: event.location,
      max_attendees: event.maxAttendees,
      price: event.price || 0,
      organizer_id: event.organizerId,
      status: 'draft'
    }),

    getUpcoming: () => {
      return this.database.query('events', {
        start_date: { $gte: new Date().toISOString() },
        status: 'published'
      }, {
        orderBy: { start_date: 'asc' }
      });
    },

    publish: (id) => this.database.update('events', id, { status: 'published' })
  };

  registrations = {
    register: (eventId, userId) => this.database.create('registrations', {
      event_id: eventId,
      user_id: userId,
      status: 'confirmed'
    }),

    getByEvent: (eventId) => {
      return this.database.query('registrations', {
        event_id: eventId
      });
    },

    checkAvailability: (eventId) => {
      return this.functions.call('check-event-availability', {
        event_id: eventId
      });
    }
  };
}
```

## Universeel Patroon

Alle deze projecten delen hetzelfde basispatroon:

1. **Extend CloudBoxUniversalClient**
2. **Create domain-specific methods** die de universele database API gebruiken
3. **Add business logic** via CloudBox functions waar nodig
4. **Use consistent data patterns** (JSON stringify voor complexe data)

## Installatie & Setup

```bash
npm install @cloudbox/universal-sdk

# Of lokaal in je project
npm install ./src/services/cloudbox-universal.ts
```

```javascript
// Voor elk project type
import { CloudBoxUniversalClient } from '@cloudbox/universal-sdk';

// Of gebruik een pre-built wrapper
import { EcommerceAPI, BlogAPI, SaaSAPI } from '@cloudbox/project-templates';

const api = new EcommerceAPI({
  endpoint: process.env.CLOUDBOX_ENDPOINT,
  apiKey: process.env.CLOUDBOX_API_KEY,
  projectId: process.env.CLOUDBOX_PROJECT_ID
});
```

Dit maakt CloudBox echt bruikbaar voor elk type project! 🚀