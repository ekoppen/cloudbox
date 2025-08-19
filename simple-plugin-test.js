const puppeteer = require('puppeteer');

async function runSimpleTest() {
    const browser = await puppeteer.launch({ 
        headless: false, 
        defaultViewport: { width: 1280, height: 800 }
    });
    
    try {
        const page = await browser.newPage();
        let jsErrors = [];
        let jsWarnings = [];
        
        // Track JavaScript errors
        page.on('console', msg => {
            if (msg.type() === 'error') {
                jsErrors.push(msg.text());
                console.log('🚨 JavaScript Error:', msg.text());
            } else if (msg.type() === 'warning') {
                jsWarnings.push(msg.text());
                console.log('⚠️ JavaScript Warning:', msg.text());
            }
        });
        
        page.on('pageerror', error => {
            jsErrors.push(error.message);
            console.log('🚨 Page Error:', error.message);
        });
        
        console.log('🔐 Starting simple plugin interface testing...');
        
        // Login first
        await page.goto('http://localhost:3000/login');
        await page.waitForSelector('input[type="email"]', { timeout: 10000 });
        
        await page.type('input[type="email"]', 'admin@cloudbox.dev');
        await page.type('input[type="password"]', 'admin123');
        await page.click('button[type="submit"]');
        
        // Wait for redirect after login
        await page.waitForNavigation({ waitUntil: 'networkidle0' });
        console.log('✅ Successfully logged in');
        
        // Test 1: Admin Plugin Interface
        console.log('\n📋 TESTING ADMIN PLUGIN INTERFACE...');
        await page.goto('http://localhost:3000/admin/plugins');
        await new Promise(resolve => setTimeout(resolve, 3000)); // Wait for page to load
        
        // Take screenshot
        await page.screenshot({ path: '/Users/eelko/Documents/_dev/cloudbox/test-admin-plugins-simple.png', fullPage: true });
        console.log('📸 Screenshot saved: test-admin-plugins-simple.png');
        
        // Check page content
        const pageContent = await page.content();
        const hasMarketplace = pageContent.includes('Marketplace') || pageContent.includes('marketplace');
        const hasPlugins = pageContent.includes('plugin') || pageContent.includes('Plugin');
        
        console.log(`📦 Page contains marketplace references: ${hasMarketplace}`);
        console.log(`📦 Page contains plugin references: ${hasPlugins}`);
        
        // Test 2: Project Plugin Interface
        console.log('\n📋 TESTING PROJECT PLUGIN INTERFACE...');
        await page.goto('http://localhost:3000/dashboard/projects/1/settings/plugins');
        await new Promise(resolve => setTimeout(resolve, 3000)); // Wait for page to load
        
        // Take screenshot
        await page.screenshot({ path: '/Users/eelko/Documents/_dev/cloudbox/test-project-plugins-simple.png', fullPage: true });
        console.log('📸 Screenshot saved: test-project-plugins-simple.png');
        
        // Check page content
        const projectPageContent = await page.content();
        const projectHasMarketplace = projectPageContent.includes('Marketplace') || projectPageContent.includes('marketplace');
        const projectHasPlugins = projectPageContent.includes('plugin') || projectPageContent.includes('Plugin');
        
        console.log(`📦 Project page contains marketplace references: ${projectHasMarketplace}`);
        console.log(`📦 Project page contains plugin references: ${projectHasPlugins}`);
        
        // Final summary
        console.log('\n🎉 SIMPLE TEST COMPLETED!');
        console.log('\n📊 ERROR SUMMARY:');
        console.log(`🚨 JavaScript Errors: ${jsErrors.length}`);
        console.log(`⚠️ JavaScript Warnings: ${jsWarnings.length}`);
        
        if (jsErrors.length > 0) {
            console.log('\n🚨 JavaScript Errors Found:');
            jsErrors.forEach((error, index) => {
                console.log(`${index + 1}. ${error}`);
            });
        }
        
        if (jsWarnings.length > 0) {
            console.log('\n⚠️ JavaScript Warnings Found:');
            jsWarnings.forEach((warning, index) => {
                console.log(`${index + 1}. ${warning}`);
            });
        }
        
        // Keep browser open for 5 seconds to see the result
        console.log('\n🔍 Keeping browser open for 5 seconds for inspection...');
        await new Promise(resolve => setTimeout(resolve, 5000));
        
    } catch (error) {
        console.error('❌ Test failed:', error);
        throw error;
    } finally {
        await browser.close();
    }
}

runSimpleTest().catch(console.error);