export const functions = {
    getAllPosts: {
        handler: "src/functions/GetAllPosts.handler",
        events: [
            {
                http: {
                    method: "GET",
                    path: "/posts",
                    cors: false,
                },
            },
        ],
    },
    getPostBySlug: {
        handler: "src/functions/GetPostBySlug.handler",
        events: [
            {
                http: {
                    method: "GET",
                    path: "/post/{slug}",
                    cors: false,
                },
            },
        ],
    },
};
