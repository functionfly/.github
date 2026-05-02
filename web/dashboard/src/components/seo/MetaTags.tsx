import { Title, Meta, OpenGraph } from 'react-meta-seo';

interface MetaTagsProps {
  title?: string;
  description?: string;
  image?: string;
  url?: string;
  type?: 'website' | 'article';
  keywords?: string[];
  author?: string;
}

export function MetaTags({
  title = "FunctionFly - Deploy Functions Anywhere",
  description = "Deploy serverless functions to any cloud provider with FunctionFly. Zero-config deployments, instant scaling, and unified developer experience.",
  image = "/og-image.svg",
  url,
  type = "website",
  keywords = ["serverless", "functions", "deployment", "cloud", "devops"],
  author = "FunctionFly"
}: MetaTagsProps) {
  return (
    <>
      <Title>{title}</Title>
      <Meta name="description" content={description} />
      <Meta name="keywords" content={Array.isArray(keywords) ? keywords.join(', ') : keywords} />
      <Meta name="author" content={author} />
      <Meta name="robots" content="index, follow" />
      <Meta name="language" content="English" />
      {url && (
        <OpenGraph
          title={title}
          description={description}
          image={image}
          url={url}
          type={type}
        />
      )}
    </>
  );
}