const { chromium } = require('playwright');

async function testPluginInterfaces() {
    console.log('🔌 Testing Plugin Interfaces...\n');
    
    const browser = await chromium.launch({ 
        headless: false,
        slowMo: 1000
    });
    
    const page = await browser.newPage();
    
    // Enable console logging
    page.on('console', msg => {
        if (msg.type() === 'error') {
            console.log('❌ Frontend Error:', msg.text());
        }
    });
    
    try {
        console.log('🔗 1. Login and navigate to dashboard...');
        await page.goto('http://localhost:3000/login');
        await page.fill('#email', 'admin@cloudbox.dev');
        await page.fill('#password', 'admin123');
        await page.click('button[type="submit"]');
        await page.waitForLoadState('networkidle');
        
        // Test Admin Plugin Interface
        console.log('🛡️ 2. Testing Admin Plugin Interface...');
        await page.goto('http://localhost:3000/dashboard/admin/plugins');
        await page.waitForLoadState('networkidle');
        
        console.log('   📸 Taking screenshot of admin plugin interface...');
        await page.screenshot({ path: 'admin-plugins.png', fullPage: true });
        
        // Check if Browse Marketplace button exists
        const marketplaceButton = page.locator('button:has-text("Browse Marketplace")');
        if (await marketplaceButton.isVisible()) {
            console.log('   ✅ Found "Browse Marketplace" button');
            
            // Click marketplace button
            await marketplaceButton.click();
            await page.waitForTimeout(2000);
            
            console.log('   📸 Taking screenshot of opened marketplace...');
            await page.screenshot({ path: 'admin-marketplace-modal.png', fullPage: true });
            
            // Test search functionality
            const searchInput = page.locator('input[placeholder*="search"], input[placeholder*="Search"]');
            if (await searchInput.isVisible()) {
                console.log('   ✅ Found search input in marketplace');
                await searchInput.fill('script');
                await page.waitForTimeout(3000);
                
                console.log('   📸 Taking screenshot of search results...');
                await page.screenshot({ path: 'admin-search-results.png', fullPage: true });
                
                // Count results
                const results = await page.locator('.plugin, [data-test*="plugin"], .marketplace').count();
                console.log(`   📊 Found ${results} search results`);
            } else {
                console.log('   ❌ Search input not found in marketplace');
            }
            
            // Close marketplace modal
            const closeButton = page.locator('button:has-text("x"), button[aria-label="close"], .close');
            if (await closeButton.first().isVisible()) {
                await closeButton.first().click();
                await page.waitForTimeout(1000);
            }
        } else {
            console.log('   ❌ "Browse Marketplace" button not found');
        }
        
        // Test Project Plugin Interface
        console.log('📁 3. Testing Project Plugin Interface...');
        
        // First get a project ID
        await page.goto('http://localhost:3000/dashboard/projects');
        await page.waitForLoadState('networkidle');
        
        // Look for a project to test with
        const projectLinks = page.locator('a[href*="/dashboard/projects/"]:has-text("Beheren")');
        const projectCount = await projectLinks.count();
        
        if (projectCount > 0) {
            console.log(`   📂 Found ${projectCount} projects, testing with first one...`);
            
            // Click on first project
            await projectLinks.first().click();
            await page.waitForLoadState('networkidle');
            
            // Navigate to project settings/plugins
            const currentUrl = page.url();
            const projectId = currentUrl.split('/').pop();
            const pluginUrl = `${currentUrl}/settings/plugins`;
            
            console.log(`   🔗 Navigating to project plugins: ${pluginUrl}`);
            await page.goto(pluginUrl);
            await page.waitForLoadState('networkidle');
            
            console.log('   📸 Taking screenshot of project plugin interface...');
            await page.screenshot({ path: 'project-plugins.png', fullPage: true });
            
            // Test marketplace tab
            const marketplaceTab = page.locator('button:has-text("Marketplace")');
            if (await marketplaceTab.isVisible()) {
                console.log('   ✅ Found Marketplace tab');
                await marketplaceTab.click();
                await page.waitForTimeout(2000);
                
                console.log('   📸 Taking screenshot of project marketplace tab...');
                await page.screenshot({ path: 'project-marketplace-tab.png', fullPage: true });
                
                // Count available plugins
                const pluginCards = await page.locator('.plugin, [data-test*="plugin"]').count();
                console.log(`   📊 Found ${pluginCards} available plugins`);
            }
            
            // Test installed tab
            const installedTab = page.locator('button:has-text("Geïnstalleerd")');
            if (await installedTab.isVisible()) {
                console.log('   ✅ Found Installed tab');
                await installedTab.click();
                await page.waitForTimeout(2000);
                
                console.log('   📸 Taking screenshot of installed plugins tab...');
                await page.screenshot({ path: 'project-installed-tab.png', fullPage: true });
                
                // Count installed plugins
                const installedPlugins = await page.locator('.plugin, [data-test*="plugin"]').count();
                console.log(`   📊 Found ${installedPlugins} installed plugins`);
            }
        } else {
            console.log('   ❌ No projects found to test with');
        }
        
        console.log('✅ 4. Testing completed!');
        
        // Wait for visual inspection
        await page.waitForTimeout(5000);
        
    } catch (error) {
        console.error('❌ Test failed:', error);
        await page.screenshot({ path: 'test-error.png', fullPage: true });
    } finally {
        await browser.close();
    }
}

testPluginInterfaces().catch(console.error);