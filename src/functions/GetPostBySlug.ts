import { response } from '@libs/api-gateway';
import { PostService } from '../libs/post.service';
import { APIGatewayEvent, APIGatewayProxyHandler } from "aws-lambda";

export const handler: APIGatewayProxyHandler = async (
  event: APIGatewayEvent
): Promise<any> => {

  if (!event.pathParameters.slug) {
    throw new Error('slug is not defined')
  }

  const postService: PostService = new PostService();
  const post = await postService.getPostBySlug(event.pathParameters.slug);
  
  return response(200, post);
};