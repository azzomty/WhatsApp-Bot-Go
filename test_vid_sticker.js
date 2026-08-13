const { Sticker } = require('wa-sticker-formatter');
const fs = require('fs');
async function test() {
    try {
        const buffer = fs.readFileSync('/tmp/test.mp4');
        const sticker = new Sticker(buffer, { pack: "Test", author: "Test" });
        await sticker.toBuffer();
        console.log("Success");
    } catch (e) {
        console.error(e);
    }
}
test();
