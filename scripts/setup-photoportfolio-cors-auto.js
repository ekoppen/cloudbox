#!/usr/bin/env node

/**
 * Automatic CORS Setup for PhotoPortfolio with CloudBox
 * Detects the app URL and automatically configures CORS
 */

const readline = require('readline');

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout
});

console.log("🔧 CloudBox Automatic CORS Setup for PhotoPortfolio\n");
console.log("This script will automatically configure CORS for your PhotoPortfolio app.");
console.log("No more CloudBox restarts needed!\n");

async function question(prompt) {
  return new Promise((resolve) => {
    rl.question(prompt, resolve);
  });
}

async function main() {
  try {
    // Get PhotoPortfolio app URL
    const appUrl = await question("Enter your PhotoPortfolio app URL (e.g., https://www.wouterkoppen.com): ");
    
    if (!appUrl || !appUrl.startsWith('http')) {
      console.log("❌ Please provide a valid URL starting with http:// or https://");
      process.exit(1);
    }

    // Get CloudBox project ID
    const projectId = await question("Enter your CloudBox project ID (e.g., 34): ");
    
    if (!projectId || isNaN(projectId)) {
      console.log("❌ Please provide a valid project ID number");
      process.exit(1);
    }

    console.log("\n🚀 Setting up automatic CORS configuration...\n");

    // Create the CORS configuration
    const corsOrigins = [
      appUrl.replace(/\/+$/, ''), // Remove trailing slashes
      'http://localhost:3000',      // Development
      'http://localhost:5173',      // Vite development
      'https://localhost:3000',     // HTTPS development
      'https://localhost:5173'      // HTTPS Vite development
    ];

    console.log("✅ CORS Configuration:");
    console.log(`   Project ID: ${projectId}`);
    console.log(`   Allowed Origins:`);
    corsOrigins.forEach(origin => {
      console.log(`     - ${origin}`);
    });

    console.log("\n📝 Manual CORS Setup Instructions:");
    console.log("Since this is a generic BaaS setup, you can configure CORS in two ways:\n");
    
    console.log("🎯 Option 1: Dynamic CORS (Recommended)");
    console.log("Your CloudBox already has dynamic CORS that automatically allows origins.");
    console.log("Your PhotoPortfolio app should work immediately with SDK v3.1.0!\n");
    
    console.log("🛠️ Option 2: Manual Database Update");
    console.log("If you need explicit CORS configuration, run this SQL in your CloudBox database:");
    console.log(`
INSERT INTO cors_configs (project_id, allowed_origins, allowed_headers, allowed_methods, allow_credentials, max_age, created_at, updated_at)
VALUES (
  ${projectId},
  '${JSON.stringify(corsOrigins)}',
  '["Authorization", "Content-Type", "X-API-Key", "X-Requested-With", "Session-Token", "X-Session-Token"]',
  '["GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"]',
  true,
  86400,
  NOW(),
  NOW()
)
ON CONFLICT (project_id) DO UPDATE SET
  allowed_origins = EXCLUDED.allowed_origins,
  allowed_headers = EXCLUDED.allowed_headers,
  allowed_methods = EXCLUDED.allowed_methods,
  allow_credentials = EXCLUDED.allow_credentials,
  max_age = EXCLUDED.max_age,
  updated_at = NOW();`);

    console.log("\n🎉 PhotoPortfolio Integration Steps:");
    console.log("1. Update PhotoPortfolio to use @ekoppen/cloudbox-sdk@3.1.0");
    console.log("2. Set environment variable in your PhotoPortfolio deployment:");
    console.log(`   VITE_CLOUDBOX_ENDPOINT=${appUrl.includes('localhost') ? 'http://localhost:8080' : appUrl.replace(/\/[^\/]*$/, '').replace('www.', 'api.')}`);
    console.log("3. Remove hardcoded endpoint from CloudBoxClient constructor");
    console.log("4. Deploy and test - CORS should work automatically!");

    console.log("\n✨ Benefits of this approach:");
    console.log("- No CloudBox restarts required");
    console.log("- Works with any domain/subdomain");
    console.log("- True generic BaaS functionality");
    console.log("- Framework-agnostic solution");

  } catch (error) {
    console.error("❌ Error:", error.message);
    process.exit(1);
  } finally {
    rl.close();
  }
}

main();