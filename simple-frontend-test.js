const { chromium } = require('playwright');

async function testFrontend() {
    console.log('🚀 Testing CloudBox Frontend...\n');
    
    const browser = await chromium.launch({ headless: false, slowMo: 1000 });
    const context = await browser.newContext();
    const page = await context.newPage();
    
    try {
        // Step 1: Load homepage
        console.log('📱 Loading homepage...');
        await page.goto('http://localhost:3001');
        await page.waitForLoadState('networkidle');
        await page.screenshot({ path: 'homepage.png' });
        console.log('✅ Homepage loaded');
        
        // Step 2: Try to find login link/button
        console.log('\n🔍 Looking for login elements...');
        
        // Check various login selectors
        const loginSelectors = [
            'a[href="/login"]',
            'button:has-text("Login")',
            'a:has-text("Login")',
            'text=Login',
            '[data-testid="login"]',
            '.login-button',
            '#login'
        ];
        
        let loginElement = null;
        for (const selector of loginSelectors) {
            try {
                loginElement = await page.locator(selector).first();
                if (await loginElement.isVisible()) {
                    console.log(`✅ Found login element with selector: ${selector}`);
                    break;
                }
            } catch (e) {
                // Continue to next selector
            }
        }
        
        if (!loginElement || !(await loginElement.isVisible())) {
            console.log('❌ No login element found, trying direct navigation...');
            await page.goto('http://localhost:3001/login');
            await page.waitForLoadState('networkidle');
            await page.screenshot({ path: 'login-page.png' });
            console.log('✅ Navigated directly to login page');
        } else {
            await loginElement.click();
            await page.waitForLoadState('networkidle');
            console.log('✅ Clicked login element');
        }
        
        // Step 3: Fill login form
        console.log('\n🔐 Testing login form...');
        
        // Wait for login form elements
        await page.waitForSelector('input[type="email"], input[name="email"]', { timeout: 10000 });
        
        const emailInput = page.locator('input[type="email"], input[name="email"]').first();
        const passwordInput = page.locator('input[type="password"], input[name="password"]').first();
        const submitButton = page.locator('button[type="submit"], button:has-text("Login"), input[type="submit"]').first();
        
        await emailInput.fill('admin@cloudbox.dev');
        await passwordInput.fill('admin123');
        await page.screenshot({ path: 'login-filled.png' });
        console.log('✅ Filled login form');
        
        await submitButton.click();
        await page.waitForLoadState('networkidle');
        await page.screenshot({ path: 'after-login.png' });
        console.log('✅ Submitted login form');
        
        // Step 4: Check if redirected to dashboard
        const currentUrl = page.url();
        console.log(`🌐 Current URL: ${currentUrl}`);
        
        if (currentUrl.includes('dashboard')) {
            console.log('✅ Successfully redirected to dashboard');
            
            // Step 5: Navigate to admin plugins
            console.log('\n🔌 Testing admin plugins page...');
            await page.goto('http://localhost:3001/dashboard/admin/plugins');
            await page.waitForLoadState('networkidle');
            await page.screenshot({ path: 'admin-plugins.png' });
            
            // Look for plugin content
            const pageContent = await page.content();
            if (pageContent.includes('Script Runner') || pageContent.includes('plugin')) {
                console.log('✅ Plugin management page loaded with content');
            } else {
                console.log('⚠️ Plugin management page loaded but no plugin content found');
            }
            
            // Step 6: Test project navigation
            console.log('\n📁 Testing project navigation...');
            await page.goto('http://localhost:3001/dashboard/projects');
            await page.waitForLoadState('networkidle');
            await page.screenshot({ path: 'projects-page.png' });
            
            // Look for project links
            const projectLinks = page.locator('a[href*="/dashboard/projects/"]');
            const projectCount = await projectLinks.count();
            
            if (projectCount > 0) {
                console.log(`✅ Found ${projectCount} project(s)`);
                
                // Click first project
                await projectLinks.first().click();
                await page.waitForLoadState('networkidle');
                await page.screenshot({ path: 'project-details.png' });
                
                // Look for Scripts menu
                const scriptsLink = page.locator('a[href*="/scripts"], text=Scripts');
                if (await scriptsLink.isVisible()) {
                    console.log('✅ Scripts menu found in project');
                    
                    await scriptsLink.click();
                    await page.waitForLoadState('networkidle');
                    await page.screenshot({ path: 'scripts-page.png' });
                    console.log('✅ Scripts page loaded');
                } else {
                    console.log('❌ Scripts menu not found in project');
                }
            } else {
                console.log('❌ No projects found');
            }
            
        } else {
            console.log('❌ Not redirected to dashboard, still on:', currentUrl);
        }
        
        console.log('\n🎉 Frontend testing complete!');
        
    } catch (error) {
        console.error('❌ Frontend test failed:', error.message);
        await page.screenshot({ path: 'error-screenshot.png' });
    } finally {
        await browser.close();
    }
}

testFrontend();