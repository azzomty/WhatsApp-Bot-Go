const { Aki } = require('aki-api');

(async () => {
    try {
        const region = 'ar';
        const aki = new Aki({ region, childMode: false });
        await aki.start();
        console.log("Question:", aki.question);
        console.log("Answers:", aki.answers);
    } catch(e) {
        console.error("ERROR:");
        console.error(e.message);
    }
})();
