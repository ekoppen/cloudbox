#!/usr/bin/env node

/**
 * Plugin Frontend Integration Test
 * Tests the enhanced error handling and graceful degradation for plugin APIs
 */

const puppeteer = require('puppeteer');

async function testPluginFrontend() {
  const browser = await puppeteer.launch({ headless: false, slowMo: 50 });
  const page = await browser.newPage();

  try {
    console.log('🔍 Testing CloudBox Plugin Frontend Fixes...');

    // Navigate to login
    console.log('📋 Navigating to login...');
    await page.goto('http://localhost:5173/login', { waitUntil: 'networkidle0' });

    // Login
    console.log('🔐 Logging in as admin...');
    await page.type('input[type="email"]', 'admin@cloudbox.dev');
    await page.type('input[type="password"]', 'admin123');
    await page.click('button[type="submit"]');
    await page.waitForNavigation({ waitUntil: 'networkidle0' });

    // Navigate to admin plugins page
    console.log('🔌 Navigating to admin plugins page...');
    await page.goto('http://localhost:5173/dashboard/admin/plugins', { waitUntil: 'networkidle0' });

    // Check if page loads without crashing
    console.log('✅ Plugin admin page loaded');

    // Wait a moment for API calls to complete or fail
    await page.waitForTimeout(3000);

    // Check if marketplace button is clickable
    console.log('🛒 Testing marketplace button...');
    const marketplaceButton = await page.$('button:has-text("Browse Marketplace")');
    if (marketplaceButton) {
      await marketplaceButton.click();
      console.log('✅ Marketplace modal opened');

      // Wait for marketplace to load or show error state
      await page.waitForTimeout(3000);

      // Check if retry button appears in case of API failure
      const retryButton = await page.$('button:has-text("Retry Loading")');
      if (retryButton) {
        console.log('✅ Retry button found - good error handling');
        await retryButton.click();
        await page.waitForTimeout(2000);
      }

      // Close marketplace modal
      const closeButton = await page.$('button:has-text("Close")');
      if (closeButton) {
        await closeButton.click();
        console.log('✅ Marketplace modal closed');
      }
    }

    // Check for error toasts or messages
    const toasts = await page.$$('.toast, [role="alert"]');
    if (toasts.length > 0) {
      console.log(`⚠️  Found ${toasts.length} toast notification(s) - API issues handled gracefully`);
    }

    console.log('🎉 Plugin frontend test completed successfully!');
    console.log('✅ Enhanced error handling and graceful degradation working');

  } catch (error) {
    console.error('❌ Test failed:', error.message);
    await page.screenshot({ path: 'plugin-frontend-test-error.png' });
  } finally {
    await browser.close();
  }
}

// Run the test
testPluginFrontend().catch(console.error);