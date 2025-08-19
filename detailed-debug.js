const puppeteer = require('puppeteer');

async function detailedDebug() {
    const browser = await puppeteer.launch({ headless: false });
    
    try {
        const page = await browser.newPage();
        
        // Track all console logs
        page.on('console', msg => {
            console.log(`[${msg.type().toUpperCase()}]`, msg.text());
        });
        
        // Login
        await page.goto('http://localhost:3000/login');
        await page.waitForSelector('input[type="email"]');
        await page.type('input[type="email"]', 'admin@cloudbox.dev');
        await page.type('input[type="password"]', 'admin123');
        await page.click('button[type="submit"]');
        await page.waitForNavigation({ waitUntil: 'networkidle0' });
        
        console.log('\n=== Going to project plugins page ===');
        
        // Go to project plugins page
        await page.goto('http://localhost:3000/dashboard/projects/1/settings/plugins', { waitUntil: 'networkidle0' });
        await new Promise(resolve => setTimeout(resolve, 5000));
        
        console.log('\n=== Test completed ===');
        
    } catch (error) {
        console.error('❌ Detailed debug failed:', error);
    } finally {
        await browser.close();
    }
}

detailedDebug().catch(console.error);