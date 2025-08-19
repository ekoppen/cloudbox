const puppeteer = require('puppeteer');

async function debugRequests() {
    const browser = await puppeteer.launch({ 
        headless: false, 
        defaultViewport: { width: 1280, height: 800 }
    });
    
    try {
        const page = await browser.newPage();
        
        // Intercept all network requests
        page.on('request', request => {
            console.log('REQUEST:', request.method(), request.url());
        });
        
        page.on('response', response => {
            console.log('RESPONSE:', response.status(), response.url());
            if (!response.ok()) {
                console.log('❌ Failed response:', response.status(), response.statusText());
            }
        });
        
        // Track JavaScript errors with more detail
        page.on('console', msg => {
            if (msg.type() === 'error') {
                console.log('🚨 JS Error:', msg.text());
                console.log('Location:', msg.location());
            }
        });
        
        page.on('pageerror', error => {
            console.log('🚨 Page Error:', error.message);
            console.log('Stack:', error.stack);
        });
        
        console.log('🔐 Starting detailed plugin debugging...');
        
        // Login first
        await page.goto('http://localhost:3000/login');
        await page.waitForSelector('input[type="email"]', { timeout: 10000 });
        
        await page.type('input[type="email"]', 'admin@cloudbox.dev');
        await page.type('input[type="password"]', 'admin123');
        await page.click('button[type="submit"]');
        
        // Wait for redirect after login
        await page.waitForNavigation({ waitUntil: 'networkidle0' });
        console.log('✅ Successfully logged in');
        
        // Go directly to project plugins page
        console.log('\n📋 Going to project plugins page...');
        await page.goto('http://localhost:3000/dashboard/projects/1/settings/plugins', { waitUntil: 'networkidle0' });
        
        // Wait for page to fully load
        await new Promise(resolve => setTimeout(resolve, 5000));
        
        // Check localStorage for auth token
        const authData = await page.evaluate(() => {
            return {
                token: localStorage.getItem('cloudbox_token'),
                user: localStorage.getItem('cloudbox_user'),
                refreshToken: localStorage.getItem('cloudbox_refresh_token'),
                allKeys: Object.keys(localStorage)
            };
        });
        console.log('🔑 Auth token found:', !!authData.token);
        console.log('👤 User data found:', !!authData.user);
        console.log('🔄 Refresh token found:', !!authData.refreshToken);
        console.log('🗂️ All localStorage keys:', authData.allKeys);
        
        // Check cookies
        const cookies = await page.cookies();
        console.log('🍪 Cookies:', cookies.map(c => c.name));
        
        // Get page title and URL
        const title = await page.title();
        const url = page.url();
        console.log('📄 Page title:', title);
        console.log('🌐 Current URL:', url);
        
        // Take final screenshot
        await page.screenshot({ path: '/Users/eelko/Documents/_dev/cloudbox/debug-plugin-requests.png', fullPage: true });
        console.log('📸 Debug screenshot saved');
        
        // Keep browser open longer for inspection
        console.log('\n🔍 Keeping browser open for 10 seconds for inspection...');
        await new Promise(resolve => setTimeout(resolve, 10000));
        
    } catch (error) {
        console.error('❌ Debug failed:', error);
    } finally {
        await browser.close();
    }
}

debugRequests().catch(console.error);