// Article 对应后端返回的文章数据结构。
export interface Article {
    ID: number;
    Title: string;
    Preview: string;
    Content: string;
    AuthorUsername: string;
    CreatedAt: string;
    UpdatedAt: string;
}

// Like 对应点赞数接口返回结构。
export interface Like{
    likes: number
}
