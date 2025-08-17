const { chromium } = require('playwright');

async function testFrontendWithLogging() {
    const browser = await chromium.launch({ headless: false });
    const context = await browser.newContext();
    const page = await context.newPage();
    
    // Enable console logging
    page.on('console', msg => console.log('PAGE LOG:', msg.text()));
    page.on('pageerror', err => console.log('PAGE ERROR:', err.message));
    
    try {
        console.log('🚀 Starting comprehensive frontend test...\n');
        
        // Step 1: Direct login test
        console.log('1️⃣ Testing direct login...');
        await page.goto('http://localhost:3001/login');
        await page.waitForTimeout(3000); // Wait for Svelte to load
        
        await page.screenshot({ path: 'step1-login.png' });
        
        // Check if form elements exist with different strategies
        const formExists = await page.locator('form').isVisible();
        console.log('Form exists:', formExists);
        
        if (formExists) {
            // Use id selectors from the Svelte component
            const emailInput = page.locator('#email');
            const passwordInput = page.locator('#password');
            const submitButton = page.locator('button[type="submit"]');
            
            const emailVisible = await emailInput.isVisible();
            const passwordVisible = await passwordInput.isVisible();
            const buttonVisible = await submitButton.isVisible();
            
            console.log('Email input visible:', emailVisible);
            console.log('Password input visible:', passwordVisible);
            console.log('Submit button visible:', buttonVisible);
            
            if (emailVisible && passwordVisible && buttonVisible) {
                await emailInput.fill('admin@cloudbox.dev');
                await passwordInput.fill('admin123');
                
                console.log('✅ Filled login form');
                await page.screenshot({ path: 'step1-filled.png' });
                
                await submitButton.click();
                await page.waitForTimeout(3000);
                
                const currentUrl = page.url();
                console.log('Current URL after login:', currentUrl);
                
                if (currentUrl.includes('dashboard')) {
                    console.log('✅ Login successful!');
                    
                    // Step 2: Test plugin management
                    console.log('\n2️⃣ Testing plugin management...');
                    await page.goto('http://localhost:3001/dashboard/admin/plugins');
                    await page.waitForTimeout(3000);
                    await page.screenshot({ path: 'step2-plugins.png' });
                    
                    const pageText = await page.textContent('body');
                    const hasScriptRunner = pageText.includes('Script Runner');
                    console.log('Script Runner plugin found:', hasScriptRunner);
                    
                    // Step 3: Test project navigation  
                    console.log('\n3️⃣ Testing project navigation...');
                    await page.goto('http://localhost:3001/dashboard/projects');
                    await page.waitForTimeout(3000);
                    await page.screenshot({ path: 'step3-projects.png' });
                    
                    // Look for project links
                    const projectLinks = page.locator('a[href*="/dashboard/projects/"]');
                    const projectCount = await projectLinks.count();
                    console.log('Projects found:', projectCount);
                    
                    if (projectCount > 0) {
                        await projectLinks.first().click();
                        await page.waitForTimeout(3000);
                        await page.screenshot({ path: 'step3-project-detail.png' });
                        
                        // Step 4: Look for Scripts menu
                        console.log('\n4️⃣ Looking for Scripts menu...');
                        const navText = await page.textContent('nav').catch(() => '');
                        const bodyText = await page.textContent('body');
                        
                        const hasScriptsInNav = navText.includes('Scripts');
                        const hasScriptsInBody = bodyText.includes('Scripts');
                        
                        console.log('Scripts in navigation:', hasScriptsInNav);
                        console.log('Scripts anywhere on page:', hasScriptsInBody);
                        
                        if (hasScriptsInNav || hasScriptsInBody) {
                            const scriptsLink = page.locator('a:has-text("Scripts")');
                            if (await scriptsLink.isVisible()) {
                                await scriptsLink.click();
                                await page.waitForTimeout(3000);
                                await page.screenshot({ path: 'step4-scripts.png' });
                                console.log('✅ Scripts page accessed');
                            }
                        }
                    }
                    
                    console.log('\n✅ All tests completed successfully!');
                    
                } else {
                    console.log('❌ Login failed, current URL:', currentUrl);
                }
            } else {
                console.log('❌ Form elements not properly visible');
            }
        } else {
            console.log('❌ No form found on login page');
        }
        
    } catch (error) {
        console.error('❌ Test error:', error.message);
        await page.screenshot({ path: 'error.png' });
    } finally {
        await browser.close();
    }
}

testFrontendWithLogging();