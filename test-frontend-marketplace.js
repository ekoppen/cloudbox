const { chromium } = require('playwright');

async function testMarketplace() {
    console.log('🌐 Testing Frontend Plugin Marketplace...\n');
    
    const browser = await chromium.launch({ 
        headless: false,
        slowMo: 1000  // Slow down for better visibility
    });
    
    try {
        const page = await browser.newPage();
        
        // Enable console logging
        page.on('console', msg => {
            if (msg.type() === 'error') {
                console.log('❌ Frontend Error:', msg.text());
            } else if (msg.type() === 'warn') {
                console.log('⚠️ Frontend Warning:', msg.text());
            }
        });
        
        // Enable network error monitoring
        page.on('response', response => {
            if (!response.ok()) {
                console.log(`❌ Network Error: ${response.status()} ${response.statusText()} - ${response.url()}`);
            }
        });
        
        console.log('🔗 Navigating to CloudBox...');
        await page.goto('http://localhost:3000');
        
        // Wait for page to load
        await page.waitForLoadState('networkidle');
        
        console.log('🔐 Logging in as admin...');
        
        // Fill in login credentials
        await page.fill('input[name="email"], input[type="email"]', 'admin@cloudbox.dev');
        await page.fill('input[name="password"], input[type="password"]', 'admin123');
        
        // Click login button
        await page.click('button[type="submit"], button:has-text("Login"), button:has-text("Sign in")');
        
        // Wait for login to complete
        await page.waitForLoadState('networkidle');
        
        console.log('📊 Navigating to admin dashboard...');
        
        // Look for admin or plugins navigation
        try {
            // Try different possible navigation paths
            const possibleAdminLinks = [
                'a:has-text("Admin")',
                'a:has-text("Dashboard")', 
                'a[href*="admin"]',
                'a[href*="dashboard"]'
            ];
            
            let adminFound = false;
            for (const selector of possibleAdminLinks) {
                try {
                    await page.click(selector, { timeout: 2000 });
                    adminFound = true;
                    break;
                } catch (e) {
                    continue;
                }
            }
            
            if (!adminFound) {
                console.log('⚠️ Could not find admin navigation, checking current page for plugin options...');
            }
            
            await page.waitForLoadState('networkidle');
            
        } catch (error) {
            console.log('⚠️ Navigation issue:', error.message);
        }
        
        console.log('🔌 Looking for plugin management...');
        
        // Look for plugin-related links or buttons
        const possiblePluginSelectors = [
            'a:has-text("Plugin")',
            'button:has-text("Plugin")',
            'a:has-text("Marketplace")',
            'button:has-text("Marketplace")',
            '[data-test="plugins"]',
            '.plugin',
            '.marketplace'
        ];
        
        let pluginSectionFound = false;
        for (const selector of possiblePluginSelectors) {
            try {
                const element = await page.locator(selector).first();
                if (await element.isVisible()) {
                    console.log(`✅ Found plugin element: ${selector}`);
                    await element.click();
                    pluginSectionFound = true;
                    break;
                }
            } catch (e) {
                continue;
            }
        }
        
        if (!pluginSectionFound) {
            console.log('❌ Could not find plugin management section');
            console.log('🔍 Current page title:', await page.title());
            console.log('🔍 Current URL:', page.url());
            
            // Take a screenshot for debugging
            await page.screenshot({ path: 'debug-current-page.png', fullPage: true });
            console.log('📸 Screenshot saved as debug-current-page.png');
        } else {
            await page.waitForLoadState('networkidle');
            
            console.log('🛒 Testing marketplace functionality...');
            
            // Look for marketplace or search functionality
            try {
                // Try to find search input
                const searchInput = page.locator('input[placeholder*="search"], input[placeholder*="Search"]').first();
                if (await searchInput.isVisible()) {
                    console.log('✅ Found search input');
                    await searchInput.fill('script');
                    await page.waitForTimeout(2000); // Wait for search to process
                    
                    // Check for results
                    const hasResults = await page.locator('.plugin, .marketplace, [data-test*="plugin"]').count() > 0;
                    if (hasResults) {
                        console.log('✅ Search results displayed');
                    } else {
                        console.log('❌ No search results found');
                    }
                } else {
                    console.log('⚠️ Search input not found');
                }
            } catch (error) {
                console.log('❌ Search test failed:', error.message);
            }
            
            // Take a screenshot of the plugin interface
            await page.screenshot({ path: 'debug-plugin-interface.png', fullPage: true });
            console.log('📸 Plugin interface screenshot saved as debug-plugin-interface.png');
        }
        
        console.log('✅ Frontend test completed');
        
    } catch (error) {
        console.error('❌ Test failed:', error);
        await page.screenshot({ path: 'debug-error.png', fullPage: true });
        console.log('📸 Error screenshot saved as debug-error.png');
    } finally {
        await browser.close();
    }
}

// Check if Playwright is available
try {
    testMarketplace().catch(console.error);
} catch (error) {
    console.log('❌ Playwright not available. Please install with: npm install playwright');
    console.log('Error:', error.message);
}