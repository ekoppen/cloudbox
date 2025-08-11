# CloudBox Universal SDK - BaaS voor Alle Projecten

## Visie: CloudBox als Echte BaaS

CloudBox moet werken zoals Firebase, Supabase, of Appwrite - een universele backend voor elk type project.

## Universele SDK Architectuur

### Core Client (Project-agnostic)
```javascript
class CloudBoxClient {
  constructor(config) {
    this.config = config;
    this.baseUrl = config.endpoint;
  }

  // Universele Database API
  database = {
    // Collections Management
    createCollection: async (name, schema) => {
      return this.request('/collections', {
        method: 'POST',
        body: JSON.stringify({ name, schema })
      });
    },

    getCollections: async () => {
      return this.request('/collections');
    },

    // Documents CRUD (werkt voor ELKE collection)
    create: async (collection, data) => {
      return this.request(`/documents/${collection}`, {
        method: 'POST',
        body: JSON.stringify(data)
      });
    },

    findMany: async (collection, options = {}) => {
      const query = this.buildQuery(options);
      return this.request(`/documents/${collection}${query}`);
    },

    findOne: async (collection, id) => {
      return this.request(`/documents/${collection}/${id}`);
    },

    update: async (collection, id, data) => {
      return this.request(`/documents/${collection}/${id}`, {
        method: 'PUT',
        body: JSON.stringify(data)
      });
    },

    delete: async (collection, id) => {
      return this.request(`/documents/${collection}/${id}`, {
        method: 'DELETE'
      });
    },

    // Advanced queries
    query: async (collection, filters, options = {}) => {
      return this.request(`/documents/${collection}`, {
        method: 'POST',
        body: JSON.stringify({ filters, options })
      });
    }
  };

  // Storage API (universeel)
  storage = {
    createBucket: async (name, options = {}) => {
      return this.request('/storage/buckets', {
        method: 'POST',
        body: JSON.stringify({
          name,
          description: options.description,
          is_public: options.isPublic,
          max_file_size: options.maxFileSize,
          allowed_mime_types: options.allowedMimeTypes
        })
      });
    },

    upload: async (bucket, file, options = {}) => {
      const formData = new FormData();
      formData.append('file', file);
      if (options.fileName) formData.append('filename', options.fileName);
      if (options.path) formData.append('path', options.path);
      
      return this.request(`/storage/${bucket}/files`, {
        method: 'POST',
        body: formData
      });
    }
  };

  // Auth API (universeel)
  auth = {
    register: async (email, password, userData = {}) => {
      return this.request('/auth/register', {
        method: 'POST',
        body: JSON.stringify({ email, password, ...userData })
      });
    },

    login: async (email, password) => {
      return this.request('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password })
      });
    },

    logout: async () => {
      return this.request('/auth/logout', { method: 'POST' });
    }
  };

  // Functions API (serverless)
  functions = {
    call: async (functionName, data = {}) => {
      return this.request(`/functions/${functionName}`, {
        method: 'POST',
        body: JSON.stringify(data)
      });
    }
  };
}
```

## Project-Specific Wrappers

### E-commerce Project
```javascript
class EcommerceAPI extends CloudBoxClient {
  // Products
  products = {
    getAll: () => this.database.findMany('products', { 
      where: { status: 'active' },
      orderBy: { created_at: 'desc' }
    }),
    
    create: (product) => this.database.create('products', {
      name: product.name,
      description: product.description,
      price: product.price,
      category_id: product.categoryId,
      images: JSON.stringify(product.images || []),
      status: 'active'
    }),

    update: (id, product) => this.database.update('products', id, product),
    delete: (id) => this.database.delete('products', id)
  };

  // Orders
  orders = {
    create: (order) => this.database.create('orders', {
      user_id: order.userId,
      items: JSON.stringify(order.items),
      total: order.total,
      status: 'pending'
    }),

    getByUser: (userId) => this.database.query('orders', {
      user_id: userId
    }),

    updateStatus: (id, status) => this.database.update('orders', id, { status })
  };

  // Categories
  categories = {
    getAll: () => this.database.findMany('categories'),
    create: (category) => this.database.create('categories', category)
  };
}
```

### Blog Project
```javascript
class BlogAPI extends CloudBoxClient {
  // Posts
  posts = {
    getPublished: () => this.database.query('posts', {
      status: 'published'
    }, {
      orderBy: { published_at: 'desc' }
    }),

    create: (post) => this.database.create('posts', {
      title: post.title,
      content: post.content,
      excerpt: post.excerpt,
      author_id: post.authorId,
      tags: JSON.stringify(post.tags || []),
      status: 'draft'
    }),

    publish: (id) => this.database.update('posts', id, {
      status: 'published',
      published_at: new Date().toISOString()
    })
  };

  // Comments
  comments = {
    getByPost: (postId) => this.database.query('comments', {
      post_id: postId,
      status: 'approved'
    }),

    create: (comment) => this.database.create('comments', {
      post_id: comment.postId,
      author_name: comment.authorName,
      author_email: comment.authorEmail,
      content: comment.content,
      status: 'pending'
    })
  };
}
```

### Portfolio Project (Bestaand)
```javascript
class PortfolioAPI extends CloudBoxClient {
  pages = {
    getPublished: () => this.database.query('pages', {
      published: 'true',
      language: 'en'
    }),

    create: (page) => this.database.create('pages', {
      title: page.title,
      slug: this.generateSlug(page.title),
      content: JSON.stringify(page.content),
      type: page.type,
      status: 'active',
      language: 'en',
      published: 'true'
    })
  };
}
```

### SaaS Application
```javascript
class SaaSAPI extends CloudBoxClient {
  // Organizations
  organizations = {
    create: (org) => this.database.create('organizations', org),
    getByUser: (userId) => this.database.query('user_organizations', {
      user_id: userId
    })
  };

  // Subscriptions
  subscriptions = {
    create: (sub) => this.database.create('subscriptions', sub),
    getByOrg: (orgId) => this.database.query('subscriptions', {
      organization_id: orgId
    })
  };
}
```

## SDK Generator Tool

Laten we een tool maken die automatisch SDK's genereert:

```javascript
// sdk-generator.js
class CloudBoxSDKGenerator {
  generateSDK(projectType, collections) {
    const template = this.getTemplate(projectType);
    const sdk = this.processTemplate(template, collections);
    return sdk;
  }

  getTemplate(projectType) {
    const templates = {
      'ecommerce': ecommerceTemplate,
      'blog': blogTemplate,
      'portfolio': portfolioTemplate,
      'saas': saasTemplate,
      'custom': genericTemplate
    };
    return templates[projectType] || templates.custom;
  }

  processTemplate(template, collections) {
    // Generate API methods based on collections
    return collections.map(collection => {
      return this.generateCRUDMethods(collection);
    });
  }

  generateCRUDMethods(collection) {
    return `
      ${collection.name} = {
        getAll: () => this.database.findMany('${collection.name}'),
        getById: (id) => this.database.findOne('${collection.name}', id),
        create: (data) => this.database.create('${collection.name}', data),
        update: (id, data) => this.database.update('${collection.name}', id, data),
        delete: (id) => this.database.delete('${collection.name}', id)
      };
    `;
  }
}
```

## Universele CLI Tool

```bash
# CloudBox project initialisatie
cloudbox init --type=ecommerce --name=my-shop
cloudbox init --type=blog --name=my-blog  
cloudbox init --type=portfolio --name=my-portfolio
cloudbox init --type=saas --name=my-app

# Collection management
cloudbox collection create products --schema=schema.json
cloudbox collection list
cloudbox collection delete products

# Data seeding
cloudbox seed --file=data.json --collection=products

# Environment management
cloudbox env set CLOUDBOX_API_KEY=abc123
cloudbox env list

# Deployment
cloudbox deploy --env=production
```

## Configuratie Templates

### E-commerce Template
```json
{
  "projectType": "ecommerce",
  "collections": {
    "products": {
      "schema": ["name", "description", "price", "category_id", "images", "status"],
      "indexes": ["category_id", "status"]
    },
    "categories": {
      "schema": ["name", "description", "parent_id"]
    },
    "orders": {
      "schema": ["user_id", "items", "total", "status", "created_at"]
    },
    "users": {
      "schema": ["email", "name", "address", "phone"]
    }
  },
  "storage": {
    "buckets": ["product-images", "user-avatars"]
  },
  "functions": ["process-payment", "send-order-confirmation"]
}
```

### Blog Template
```json
{
  "projectType": "blog",
  "collections": {
    "posts": {
      "schema": ["title", "content", "excerpt", "author_id", "tags", "status", "published_at"]
    },
    "authors": {
      "schema": ["name", "bio", "avatar", "social_links"]
    },
    "comments": {
      "schema": ["post_id", "author_name", "author_email", "content", "status"]
    }
  },
  "storage": {
    "buckets": ["post-images", "author-avatars"]
  }
}
```

## Migration Path

Voor bestaande projecten:

```javascript
// Oude manier (portfolio-specific)
const pages = await cloudbox.portfolio.getPages();

// Nieuwe manier (universeel)
const pages = await cloudbox.database.query('pages', {
  published: 'true',
  language: 'en'
});

// Of met wrapper
const portfolio = new PortfolioAPI(config);
const pages = await portfolio.pages.getPublished();
```

## Voordelen van deze Aanpak

1. **Universeel**: Werkt voor elk project type
2. **Consistent**: Dezelfde API patterns voor alle projecten  
3. **Schaalbaar**: Makkelijk uit te breiden met nieuwe project types
4. **Developer-friendly**: Type-safe met TypeScript support
5. **Tool ecosystem**: CLI, generators, templates

Deze aanpak maakt CloudBox een echte concurrent voor Firebase/Supabase! 🚀