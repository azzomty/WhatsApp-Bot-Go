const puppeteer = require('puppeteer-extra');
const StealthPlugin = require('puppeteer-extra-plugin-stealth');
puppeteer.use(StealthPlugin());

(async () => {
    const browser = await puppeteer.launch({
        executablePath: '/usr/bin/google-chrome',
        headless: "new",
        args: ['--no-sandbox', '--disable-setuid-sandbox']
    });
    const page = await browser.newPage();
    
    await page.goto('https://ar.akinator.com/', { waitUntil: 'domcontentloaded' });
    
    await Promise.all([
        page.evaluate(() => document.getElementById('formTheme').submit()),
        page.waitForNavigation({ waitUntil: 'domcontentloaded' })
    ]);
    
    await Promise.all([
        page.evaluate(() => document.querySelector('.li-game').click()),
        page.waitForNavigation({ waitUntil: 'domcontentloaded' })
    ]);
    
    try {
        await page.waitForSelector('.question-text', { timeout: 15000 });
        const question = await page.$eval('.question-text', el => el.innerText);
        console.log("Question:", question);
        const answers = await page.$$eval('.a-base', els => els.map(el => el.innerText));
        console.log("Answers:", answers);
    } catch (e) {
        console.log("Failed to find question.");
    }
    
    await browser.close();
})();
