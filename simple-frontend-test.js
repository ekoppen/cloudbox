const { chromium } = require('playwright');

async function simpleTest() {
    console.log('🌐 Simple Frontend Test...\n');
    
    const browser = await chromium.launch({ 
        headless: false,
        slowMo: 500
    });
    
    const page = await browser.newPage();
    
    try {
        console.log('🔗 Navigating to CloudBox...');
        await page.goto('http://localhost:3000');
        
        await page.waitForLoadState('networkidle');
        
        console.log('📋 Page title:', await page.title());
        console.log('🔗 Current URL:', page.url());
        
        // Check what's on the page
        const bodyText = await page.textContent('body');
        console.log('📄 Page contains login elements:', bodyText.toLowerCase().includes('login') || bodyText.toLowerCase().includes('email'));
        
        // Take a screenshot
        await page.screenshot({ path: 'current-page.png', fullPage: true });
        console.log('📸 Screenshot saved as current-page.png');
        
        // Look for any input fields
        const inputs = await page.locator('input').count();
        console.log('📝 Number of input fields:', inputs);
        
        if (inputs > 0) {
            for (let i = 0; i < Math.min(inputs, 5); i++) {
                const input = page.locator('input').nth(i);
                const type = await input.getAttribute('type');
                const placeholder = await input.getAttribute('placeholder');
                const name = await input.getAttribute('name');
                console.log(`   Input ${i}: type="${type}", name="${name}", placeholder="${placeholder}"`);
            }
        }
        
        // Wait a bit for visual inspection
        await page.waitForTimeout(5000);
        
    } catch (error) {
        console.error('❌ Test failed:', error);
        await page.screenshot({ path: 'error-page.png', fullPage: true });
    } finally {
        await browser.close();
    }
}

simpleTest().catch(console.error);
