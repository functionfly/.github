"use client";
import { Studio as SanityStudio } from "sanity";
import config from "../../studio/sanity.config";

export default function Studio() {
  // @ts-ignore
  return <SanityStudio config={config} />;
}
