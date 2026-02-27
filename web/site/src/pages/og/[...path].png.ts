import { ImageResponse } from '@vercel/og';

export const config = {
  runtime: 'edge',
};

export async function GET({ params, request }) {
  const { path } = params;
  const url = new URL(request.url);

  // Extract title from query params or use default
  const title = url.searchParams.get('title') || 'FunctionFly';
  const description = url.searchParams.get('description') || 'Serverless Reliability Platform';

  // Load Inter font
  const fontData = await fetch(
    'https://fonts.gstatic.com/s/inter/v12/UcCO3FwrK3iLTeHuS_fvQtMwCp50KnMw2boKoduKmMEVuLyfAZ9hiJ-Ek-_EeA.woff'
  ).then(res => res.arrayBuffer());

  return new ImageResponse(
    (
      <div
        style={{
          height: '100%',
          width: '100%',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          backgroundColor: '#0a0a0f',
          backgroundImage: 'radial-gradient(circle at 25% 25%, #3b82f6 0%, transparent 50%), radial-gradient(circle at 75% 75%, #1e40af 0%, transparent 50%)',
        }}
      >
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            backgroundColor: 'rgba(255, 255, 255, 0.05)',
            borderRadius: '16px',
            padding: '40px',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            maxWidth: '900px',
            textAlign: 'center',
          }}
        >
          <h1
            style={{
              fontSize: '64px',
              fontWeight: 'bold',
              color: '#ffffff',
              margin: '0 0 20px 0',
              lineHeight: '1.1',
              fontFamily: 'Inter',
            }}
          >
            {title}
          </h1>
          <p
            style={{
              fontSize: '24px',
              color: '#94a3b8',
              margin: '0',
              lineHeight: '1.4',
              maxWidth: '600px',
              fontFamily: 'Inter',
            }}
          >
            {description}
          </p>
        </div>

        <div
          style={{
            position: 'absolute',
            bottom: '40px',
            right: '40px',
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
          }}
        >
          <div
            style={{
              width: '32px',
              height: '32px',
              backgroundColor: '#3b82f6',
              borderRadius: '50%',
            }}
          />
          <span
            style={{
              color: '#ffffff',
              fontSize: '18px',
              fontWeight: '600',
              fontFamily: 'Inter',
            }}
          >
            functionfly.com
          </span>
        </div>
      </div>
    ),
    {
      width: 1200,
      height: 630,
      fonts: [
        {
          name: 'Inter',
          data: fontData,
          weight: 600,
          style: 'normal',
        },
      ],
    }
  );
}