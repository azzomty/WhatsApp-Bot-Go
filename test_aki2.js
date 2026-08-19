const { Akinator } = require('@aqul/akinator-api');
(async () => {
    try {
        const region = 'ar';
        const aki = new Akinator(region, true); // region, childMode
        await aki.start();
        console.log("Question:", aki.question);
        console.log("Answers:", aki.answers);
    } catch(e) {
        console.error("ERROR:");
        console.error(e.message);
    }
})();
