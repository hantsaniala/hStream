import axios from "axios";

axios.defaults.baseURL = `${process.env.VUE_APP_API_URL}`;

export default class Http {
  static async get(url, params) {
    try {
      const res = await axios.get(url, {
        params,
      });
      return res;
    } catch (error) {
      console.error(error);
      throw error;
    }
  }
  static async post(url, data) {
    try {
      const res = await axios.post(url, data);
      return res;
    } catch (error) {
      console.error(error);
      throw error;
    }
  }
}
