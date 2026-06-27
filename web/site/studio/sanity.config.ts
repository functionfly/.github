import { defineConfig } from "sanity";
import { schemaTypes } from "./schemas";

export default defineConfig({
  name: "functionfly-blog",
  title: "FunctionFly Blog",

  projectId: process.env.SANITY_STUDIO_PROJECT_ID || "",
  dataset: process.env.SANITY_STUDIO_DATASET || "production",

  basePath: "/studio",

  plugins: [],

  schema: {
    types: schemaTypes,
  },

  structure: (S) =>
    S.list()
      .title("Content")
      .items([
        S.listItem()
          .title("Reports")
          .child(
            S.documentTypeList("report").title("State of AI Builders Reports"),
          ),
        S.listItem()
          .title("Blog Posts")
          .child(S.documentTypeList("blogPost").title("Blog Posts")),
        S.divider(),
        S.listItem()
          .title("Authors")
          .child(S.documentTypeList("author").title("Authors")),
        S.listItem()
          .title("Categories")
          .child(S.documentTypeList("category").title("Categories")),
      ]),
});
