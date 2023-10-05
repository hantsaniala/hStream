import Http from "./http";

export default class VideoServices extends Http {
  static async postVideo(data) {
    try {
      return await this.post("/api/v1/videos", data);
    } catch (error) {
      console.error(error);
      return error;
    }
  }
}
