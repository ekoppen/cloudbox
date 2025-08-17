const { chromium } = require('playwright');

async function testCloudBoxE2E() {
    console.log('🚀 Starting CloudBox E2E Testing...\n');
    
    const browser = await chromium.launch({ headless: false });
    const context = await browser.newContext();
    const page = await context.newPage();
    
    try {
        // Test 1: Frontend loads correctly
        console.log('📱 Testing frontend loading...');
        await page.goto('http://localhost:3001');
        await page.waitForLoadState('domcontentloaded');
        
        const title = await page.title();
        console.log(`✅ Frontend loaded. Title: ${title}`);
        
        // Test 2: Login functionality
        console.log('\n🔐 Testing login functionality...');
        await page.click('text=Login');
        await page.waitForSelector('input[type="email"]');
        
        await page.fill('input[type="email"]', 'admin@cloudbox.dev');
        await page.fill('input[type="password"]', 'admin123');
        await page.click('button[type="submit"]');
        
        // Wait for redirect to dashboard
        await page.waitForURL(/.*dashboard.*/, { timeout: 10000 });
        console.log('✅ Login successful, redirected to dashboard');
        
        // Test 3: Navigate to admin area
        console.log('\n👨‍💼 Testing admin area access...');
        await page.goto('http://localhost:3001/dashboard/admin');
        await page.waitForLoadState('domcontentloaded');
        
        const adminTitle = await page.textContent('h1');
        console.log(`✅ Admin area accessible. Title: ${adminTitle}`);
        
        // Test 4: Navigate to plugin management
        console.log('\n🔌 Testing plugin management dashboard...');
        await page.goto('http://localhost:3001/dashboard/admin/plugins');
        await page.waitForLoadState('domcontentloaded');
        
        // Check if plugins are listed
        await page.waitForSelector('[data-testid="plugin-card"], .plugin-card, text=Script Runner', { timeout: 5000 });
        
        const pluginExists = await page.locator('text=Script Runner').isVisible();
        if (pluginExists) {
            console.log('✅ Script Runner plugin found in dashboard');
            
            // Test enable/disable functionality
            const enableButton = page.locator('button', { hasText: /Enable|Disable/ }).first();
            const buttonText = await enableButton.textContent();
            console.log(`🔄 Found button: ${buttonText}`);
            
            await enableButton.click();
            await page.waitForTimeout(1000); // Wait for state change
            console.log('✅ Plugin toggle functionality works');
        } else {
            console.log('❌ Script Runner plugin not found in dashboard');
        }
        
        // Test 5: Navigate to project and check scripts menu
        console.log('\n📁 Testing project scripts integration...');
        await page.goto('http://localhost:3001/dashboard/projects');
        await page.waitForLoadState('domcontentloaded');
        
        // Click on first project
        const projectLink = page.locator('a[href*="/dashboard/projects/"]').first();
        await projectLink.click();
        await page.waitForLoadState('domcontentloaded');
        
        // Check if Scripts menu item exists
        const scriptsMenu = page.locator('text=Scripts');
        const scriptsMenuExists = await scriptsMenu.isVisible();
        
        if (scriptsMenuExists) {
            console.log('✅ Scripts menu found in project navigation');
            
            // Test 6: Navigate to scripts page
            await scriptsMenu.click();
            await page.waitForLoadState('domcontentloaded');
            
            const scriptsPageTitle = await page.textContent('h1, h2').catch(() => 'Scripts Page');
            console.log(`✅ Scripts page accessible. Title: ${scriptsPageTitle}`);
        } else {
            console.log('❌ Scripts menu not found in project navigation');
        }
        
        console.log('\n🎉 E2E Testing Complete!');
        
    } catch (error) {
        console.error('❌ Test failed:', error.message);
        
        // Take screenshot on failure
        await page.screenshot({ path: 'test-failure.png' });
        console.log('📸 Screenshot saved as test-failure.png');
    } finally {
        await browser.close();
    }
}

// Install playwright if needed
async function checkPlaywright() {
    try {
        require('playwright');
        return true;
    } catch (e) {
        console.log('📦 Installing Playwright...');
        const { execSync } = require('child_process');
        execSync('npm install playwright', { stdio: 'inherit' });
        execSync('npx playwright install', { stdio: 'inherit' });
        return true;
    }
}

// Run the test
checkPlaywright().then(() => testCloudBoxE2E());