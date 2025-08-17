const { chromium } = require('playwright');

async function debugFrontend() {
    const browser = await chromium.launch({ headless: false });
    const page = await browser.newPage();
    
    try {
        // Go to login page directly
        console.log('Going to login page...');
        await page.goto('http://localhost:3001/login');
        await page.waitForLoadState('networkidle');
        
        // Take screenshot
        await page.screenshot({ path: 'debug-login.png' });
        
        // Get page content
        const content = await page.content();
        console.log('Page content length:', content.length);
        
        // Look for form elements
        const inputs = await page.locator('input').count();
        const buttons = await page.locator('button').count();
        console.log(`Found ${inputs} inputs and ${buttons} buttons`);
        
        // Get all input types
        for (let i = 0; i < inputs; i++) {
            const input = page.locator('input').nth(i);
            const type = await input.getAttribute('type');
            const name = await input.getAttribute('name');
            const placeholder = await input.getAttribute('placeholder');
            console.log(`Input ${i}: type=${type}, name=${name}, placeholder=${placeholder}`);
        }
        
        // Wait a bit for any dynamic content
        await page.waitForTimeout(3000);
        await page.screenshot({ path: 'debug-login-after-wait.png' });
        
    } catch (error) {
        console.error('Debug failed:', error);
    } finally {
        await browser.close();
    }
}

debugFrontend();