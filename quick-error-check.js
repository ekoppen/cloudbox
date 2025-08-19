const puppeteer = require('puppeteer');

async function quickErrorCheck() {
    const browser = await puppeteer.launch({ headless: true });
    
    try {
        const page = await browser.newPage();
        
        // Track JavaScript errors with more detail
        page.on('console', msg => {
            if (msg.type() === 'error') {
                console.log('🚨 JS Error:', msg.text());
                console.log('Args:', msg.args().map(arg => arg.toString()));
            }
        });
        
        // Login
        await page.goto('http://localhost:3000/login');
        await page.waitForSelector('input[type="email"]');
        await page.type('input[type="email"]', 'admin@cloudbox.dev');
        await page.type('input[type="password"]', 'admin123');
        await page.click('button[type="submit"]');
        await page.waitForNavigation({ waitUntil: 'networkidle0' });
        
        // Go to project plugins page
        await page.goto('http://localhost:3000/dashboard/projects/1/settings/plugins', { waitUntil: 'networkidle0' });
        await new Promise(resolve => setTimeout(resolve, 3000));
        
    } catch (error) {
        console.error('❌ Quick check failed:', error);
    } finally {
        await browser.close();
    }
}

quickErrorCheck().catch(console.error);