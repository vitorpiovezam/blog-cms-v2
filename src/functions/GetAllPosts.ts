import { response } from '@libs/api-gateway';
import { PostService } from '../libs/post.service';
import { APIGatewayEvent, APIGatewayProxyHandler } from "aws-lambda";

export const handler: APIGatewayProxyHandler = async (
  event: APIGatewayEvent
): Promise<any> => {

  const postService: PostService = new PostService();
  const posts = await postService.getAllPosts();

  return response(200, posts);
};
