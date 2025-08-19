const { chromium } = require('playwright');

async function testCompleteFlow() {
    console.log('🌐 Complete Frontend Plugin Marketplace Test...\n');
    
    const browser = await chromium.launch({ 
        headless: false,
        slowMo: 1000
    });
    
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
        if (!response.ok() && response.url().includes('api')) {
            console.log(`❌ API Error: ${response.status()} ${response.statusText()} - ${response.url()}`);
        }
    });
    
    try {
        console.log('🔗 1. Navigating to CloudBox homepage...');
        await page.goto('http://localhost:3000');
        await page.waitForLoadState('networkidle');
        
        console.log('🔗 2. Going to login page...');
        await page.click('a[href="/login"]');
        await page.waitForLoadState('networkidle');
        
        console.log('🔐 3. Logging in as admin...');
        await page.fill('#email', 'admin@cloudbox.dev');
        await page.fill('#password', 'admin123');
        await page.click('button[type="submit"]');
        
        // Wait for login redirect
        await page.waitForLoadState('networkidle');
        console.log('🔗 Current URL after login:', page.url());
        
        // Take a screenshot after login
        await page.screenshot({ path: 'after-login.png', fullPage: true });
        console.log('📸 Screenshot saved as after-login.png');
        
        console.log('📊 4. Looking for admin/dashboard navigation...');
        
        // Look for navigation elements
        const navItems = await page.locator('nav a, .nav a, a:has-text("Admin"), a:has-text("Dashboard"), a:has-text("Plugin")').count();
        console.log(`Found ${navItems} navigation items`);
        
        // Try different navigation approaches
        let adminFound = false;
        
        // Try to find admin or dashboard link
        const possibleNavs = [
            'a:has-text("Admin")',
            'a:has-text("Dashboard")', 
            'a[href*="admin"]',
            'a[href*="dashboard"]',
            'button:has-text("Admin")',
            'button:has-text("Dashboard")'
        ];
        
        for (const selector of possibleNavs) {
            try {
                const element = await page.locator(selector).first();
                if (await element.isVisible()) {
                    console.log(`✅ Found navigation: ${selector}`);
                    await element.click();
                    await page.waitForLoadState('networkidle');
                    adminFound = true;
                    break;
                }
            } catch (e) {
                // Continue to next selector
            }
        }
        
        if (!adminFound) {
            console.log('⚠️ No explicit admin nav found, checking current page for plugin options...');
            console.log('🔗 Current URL:', page.url());
            
            // Look for any plugin-related elements on current page
            const pluginElements = await page.locator('[data-test*="plugin"], .plugin, a:has-text("Plugin"), button:has-text("Plugin")').count();
            console.log(`Found ${pluginElements} plugin-related elements on current page`);
        }
        
        console.log('🔌 5. Looking for plugin management interface...');
        
        // Take screenshot of current state
        await page.screenshot({ path: 'looking-for-plugins.png', fullPage: true });
        console.log('📸 Screenshot saved as looking-for-plugins.png');
        
        // Look for plugin or marketplace elements
        const pluginSelectors = [
            'a:has-text("Plugin")',
            'button:has-text("Plugin")',
            'a:has-text("Marketplace")',
            'button:has-text("Marketplace")',
            '[data-test="plugins"]',
            '.plugin-marketplace',
            '.marketplace'
        ];
        
        let pluginInterfaceFound = false;
        
        for (const selector of pluginSelectors) {
            try {
                const element = await page.locator(selector).first();
                if (await element.isVisible()) {
                    console.log(`✅ Found plugin interface: ${selector}`);
                    await element.click();
                    await page.waitForLoadState('networkidle');
                    pluginInterfaceFound = true;
                    break;
                }
            } catch (e) {
                // Continue to next selector
            }
        }
        
        if (pluginInterfaceFound) {
            console.log('🛒 6. Testing marketplace functionality...');
            
            // Take screenshot of plugin interface
            await page.screenshot({ path: 'plugin-interface.png', fullPage: true });
            console.log('📸 Plugin interface screenshot saved');
            
            // Test search functionality
            try {
                const searchInput = page.locator('input[placeholder*="search"], input[placeholder*="Search"]').first();
                if (await searchInput.isVisible()) {
                    console.log('✅ Found search input, testing search...');
                    await searchInput.fill('script');
                    await page.waitForTimeout(3000); // Wait for search to process
                    
                    // Look for results
                    const resultsContainer = page.locator('.plugin, .marketplace-plugin, [data-test*="plugin"]');
                    const resultCount = await resultsContainer.count();
                    
                    if (resultCount > 0) {
                        console.log(`✅ Search results displayed: ${resultCount} items`);
                        
                        // Take screenshot of search results
                        await page.screenshot({ path: 'search-results.png', fullPage: true });
                        console.log('📸 Search results screenshot saved');
                        
                        // Test clicking on a result
                        try {
                            await resultsContainer.first().click();
                            await page.waitForTimeout(2000);
                            console.log('✅ Clicked on first search result');
                            
                            await page.screenshot({ path: 'plugin-details.png', fullPage: true });
                            console.log('📸 Plugin details screenshot saved');
                        } catch (e) {
                            console.log('⚠️ Could not click on search result:', e.message);
                        }
                    } else {
                        console.log('❌ No search results found');
                    }
                } else {
                    console.log('❌ Search input not found');
                }
            } catch (error) {
                console.log('❌ Search functionality test failed:', error.message);
            }
        } else {
            console.log('❌ Could not find plugin management interface');
            
            // Let's check what's available on the current page
            const bodyText = await page.textContent('body');
            console.log('📄 Page contains "plugin":', bodyText.toLowerCase().includes('plugin'));
            console.log('📄 Page contains "marketplace":', bodyText.toLowerCase().includes('marketplace'));
            
            // List all links on the page
            const links = await page.locator('a').evaluateAll(links => 
                links.map(link => ({ text: link.textContent?.trim(), href: link.href }))
                     .filter(link => link.text && link.text.length > 0)
                     .slice(0, 10)
            );
            console.log('🔗 Available links:', links);
        }
        
        console.log('✅ 7. Frontend test completed');
        
        // Wait for visual inspection
        await page.waitForTimeout(5000);
        
    } catch (error) {
        console.error('❌ Test failed:', error);
        await page.screenshot({ path: 'test-error.png', fullPage: true });
        console.log('📸 Error screenshot saved as test-error.png');
    } finally {
        await browser.close();
    }
}

testCompleteFlow().catch(console.error);