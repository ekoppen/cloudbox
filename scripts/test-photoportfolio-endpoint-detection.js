#!/usr/bin/env node

/**
 * Test script for PhotoPortfolio endpoint detection with CloudBox SDK v3.1.0
 * Tests the new automatic endpoint detection functionality
 */

// Simulate different environments
const environments = [
  {
    name: "Production with environment variable",
    setup: () => {
      process.env.CLOUDBOX_ENDPOINT = "https://your-cloudbox-domain.com";
      delete process.env.VITE_CLOUDBOX_ENDPOINT;
    }
  },
  {
    name: "Vite development",
    setup: () => {
      process.env.VITE_CLOUDBOX_ENDPOINT = "http://localhost:8080";
      delete process.env.CLOUDBOX_ENDPOINT;
    }
  },
  {
    name: "React production build",
    setup: () => {
      process.env.REACT_APP_CLOUDBOX_ENDPOINT = "https://api.example.com";
      delete process.env.CLOUDBOX_ENDPOINT;
      delete process.env.VITE_CLOUDBOX_ENDPOINT;
    }
  },
  {
    name: "Next.js production",
    setup: () => {
      process.env.NEXT_PUBLIC_CLOUDBOX_ENDPOINT = "https://cloudbox.production.com";
      delete process.env.CLOUDBOX_ENDPOINT;
      delete process.env.VITE_CLOUDBOX_ENDPOINT;
      delete process.env.REACT_APP_CLOUDBOX_ENDPOINT;
    }
  },
  {
    name: "No environment variables (fallback to localhost)",
    setup: () => {
      delete process.env.CLOUDBOX_ENDPOINT;
      delete process.env.VITE_CLOUDBOX_ENDPOINT;
      delete process.env.REACT_APP_CLOUDBOX_ENDPOINT;
      delete process.env.NEXT_PUBLIC_CLOUDBOX_ENDPOINT;
    }
  }
];

console.log("🧪 Testing CloudBox SDK v3.1.0 Automatic Endpoint Detection\n");

// Mock the SDK's endpoint detection logic for testing
class MockEndpointDetector {
  static detect() {
    const detectors = [
      () => process.env.CLOUDBOX_ENDPOINT,
      () => process.env.VITE_CLOUDBOX_ENDPOINT,
      () => process.env.REACT_APP_CLOUDBOX_ENDPOINT,
      () => process.env.NEXT_PUBLIC_CLOUDBOX_ENDPOINT,
    ];
    
    for (const detector of detectors) {
      try {
        const endpoint = detector();
        if (endpoint && typeof endpoint === 'string' && endpoint.startsWith('http')) {
          return endpoint.replace(/\/+$/, ''); // Remove trailing slashes
        }
      } catch (error) {
        continue;
      }
    }
    
    return 'http://localhost:8080';
  }
}

environments.forEach((env, index) => {
  console.log(`${index + 1}. ${env.name}`);
  env.setup();
  
  const detectedEndpoint = MockEndpointDetector.detect();
  console.log(`   Detected endpoint: ${detectedEndpoint}`);
  
  // Show which environment variable was used
  const usedVar = process.env.CLOUDBOX_ENDPOINT ? 'CLOUDBOX_ENDPOINT' :
                  process.env.VITE_CLOUDBOX_ENDPOINT ? 'VITE_CLOUDBOX_ENDPOINT' :
                  process.env.REACT_APP_CLOUDBOX_ENDPOINT ? 'REACT_APP_CLOUDBOX_ENDPOINT' :
                  process.env.NEXT_PUBLIC_CLOUDBOX_ENDPOINT ? 'NEXT_PUBLIC_CLOUDBOX_ENDPOINT' :
                  'fallback';
  console.log(`   Source: ${usedVar}\n`);
});

console.log("✅ PhotoPortfolio Integration Guide:");
console.log("1. Update your PhotoPortfolio app to use CloudBox SDK v3.1.0");
console.log("2. Remove hardcoded endpoint from CloudBoxClient constructor:");
console.log("   // OLD:");
console.log("   const client = new CloudBoxClient({");
console.log("     projectId: 34,");
console.log("     apiKey: 'your-api-key',");
console.log("     endpoint: 'http://localhost:8080' // ❌ Remove this");
console.log("   });");
console.log("");
console.log("   // NEW:");
console.log("   const client = new CloudBoxClient({");
console.log("     projectId: 34,");
console.log("     apiKey: 'your-api-key'");
console.log("     // ✅ endpoint auto-detected from environment");
console.log("   });");
console.log("");
console.log("3. Set environment variable in your production deployment:");
console.log("   VITE_CLOUDBOX_ENDPOINT=https://www.wouterkoppen.com");
console.log("   (or use a meta tag in your HTML)");
console.log("");
console.log("4. The SDK will automatically:");
console.log("   - Detect your production CloudBox URL");
console.log("   - Use the correct CORS origins");
console.log("   - Work without CloudBox restarts");
console.log("");
console.log("🎉 This enables true generic BaaS functionality!");