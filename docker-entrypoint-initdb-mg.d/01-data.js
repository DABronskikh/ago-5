db.createCollection('operations');

db.operations.insertOne({
    user_id: 1,
    operations: [
        {
            icon: "https://img-1",
            title: "Организациям",
            url: "https://url-1"
        },
        {
            icon: "https://img-2",
            title: "Интернет",
            url: "https://url-2"
        },
        {
            icon: "https://img-3",
            title: "Запрос денег",
            url: "https://url-3"
        }
    ],
});

db.operations.insertOne({
    user_id: 2,
    operations: [
        {
            icon: "https://img-1",
            title: "Организациям",
            url: "https://url-1"
        },
    ],
});
