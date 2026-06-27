import { defineField, defineType } from "sanity";

export default defineType({
  name: "report",
  title: "State of AI Builders Report",
  type: "document",
  fields: [
    defineField({
      name: "title",
      title: "Title",
      type: "string",
      validation: (Rule) => Rule.required(),
    }),
    defineField({
      name: "slug",
      title: "Slug",
      type: "slug",
      options: { source: "title" },
      validation: (Rule) => Rule.required(),
    }),
    defineField({
      name: "periodStart",
      title: "Period Start",
      type: "date",
    }),
    defineField({
      name: "periodEnd",
      title: "Period End",
      type: "date",
    }),
    defineField({
      name: "generatedAt",
      title: "Generated At",
      type: "datetime",
    }),
    defineField({
      name: "tldrMetros",
      title: "TL;DR: Metros Ranked",
      type: "number",
    }),
    defineField({
      name: "tldrUniversities",
      title: "TL;DR: Universities Ranked",
      type: "number",
    }),
    defineField({
      name: "tldrActiveBuilders",
      title: "TL;DR: Active Builders",
      type: "number",
    }),
    defineField({
      name: "tldrNewBuilders",
      title: "TL;DR: New Builders",
      type: "number",
    }),
    defineField({
      name: "tldrDeployments",
      title: "TL;DR: Deployments",
      type: "number",
    }),
    defineField({
      name: "tldrExecutions",
      title: "TL;DR: Executions",
      type: "number",
    }),
    defineField({
      name: "tldrAmbassadors",
      title: "TL;DR: New Ambassadors",
      type: "number",
    }),
    defineField({
      name: "topMetros",
      title: "Top 10 Metros",
      type: "array",
      of: [
        {
          type: "object",
          name: "metroEntry",
          fields: [
            { name: "rank", title: "Rank", type: "number" },
            { name: "city", title: "City", type: "string" },
            { name: "country", title: "Country", type: "string" },
            { name: "population", title: "Population", type: "number" },
            { name: "perCapita", title: "Per Capita", type: "number" },
            { name: "builders", title: "Builders", type: "number" },
          ],
        },
      ],
    }),
    defineField({
      name: "topUniversities",
      title: "Top 10 Universities",
      type: "array",
      of: [
        {
          type: "object",
          name: "universityEntry",
          fields: [
            { name: "rank", title: "Rank", type: "number" },
            { name: "university", title: "University", type: "string" },
            { name: "country", title: "Country", type: "string" },
            { name: "students", title: "Students", type: "number" },
            { name: "perCapita", title: "Per Capita", type: "number" },
            { name: "builders", title: "Builders", type: "number" },
          ],
        },
      ],
    }),
    defineField({
      name: "newMetros",
      title: "New Metros This Month",
      type: "array",
      of: [{ type: "block" }],
    }),
    defineField({
      name: "newAmbassadors",
      title: "New Ambassadors",
      type: "array",
      of: [{ type: "block" }],
    }),
    defineField({
      name: "biggestGainers",
      title: "Biggest Gainers",
      type: "array",
      of: [
        {
          type: "object",
          name: "cityDelta",
          fields: [
            { name: "delta", title: "Δ", type: "string" },
            { name: "city", title: "City", type: "string" },
            { name: "newRank", title: "New Rank", type: "string" },
            { name: "perCapita", title: "Per Capita", type: "number" },
          ],
        },
      ],
    }),
    defineField({
      name: "biggestLosers",
      title: "Biggest Losers",
      type: "array",
      of: [
        {
          type: "object",
          name: "cityDelta",
          fields: [
            { name: "delta", title: "Δ", type: "string" },
            { name: "city", title: "City", type: "string" },
            { name: "newRank", title: "New Rank", type: "string" },
            { name: "perCapita", title: "Per Capita", type: "number" },
          ],
        },
      ],
    }),
    defineField({
      name: "body",
      title: "Full Report Body (Markdown)",
      type: "text",
    }),
  ],
  preview: {
    select: { title: "title", subtitle: "generatedAt" },
  },
});
