import { Volume2, Play, Pause } from 'lucide-react';
import { useState, useRef, useEffect } from 'react';

interface AudioPlayerProps {
  url?: string;
  path?: string;
  format?: string;
  text?: string;
  voice?: string;
}

export function AudioPlayer({ url, path, format, text, voice }: AudioPlayerProps) {
  const audioRef = useRef<HTMLAudioElement>(null);
  const [playing, setPlaying] = useState(false);
  const [duration, setDuration] = useState(0);
  const [currentTime, setCurrentTime] = useState(0);

  useEffect(() => {
    const audio = audioRef.current;
    if (!audio) return;
    const onTime = () => setCurrentTime(audio.currentTime);
    const onDur = () => setDuration(audio.duration);
    const onEnd = () => setPlaying(false);
    audio.addEventListener('timeupdate', onTime);
    audio.addEventListener('loadedmetadata', onDur);
    audio.addEventListener('ended', onEnd);
    return () => {
      audio.removeEventListener('timeupdate', onTime);
      audio.removeEventListener('loadedmetadata', onDur);
      audio.removeEventListener('ended', onEnd);
    };
  }, [url]);

  const togglePlay = () => {
    const audio = audioRef.current;
    if (!audio) return;
    if (playing) {
      audio.pause();
    } else {
      audio.play();
    }
    setPlaying(!playing);
  };

  const formatTime = (s: number) => {
    const m = Math.floor(s / 60);
    const sec = Math.floor(s % 60);
    return `${m}:${sec.toString().padStart(2, '0')}`;
  };

  return (
    <div style={{
      background: 'var(--panel-raised)', borderRadius: 'var(--radius-sm)',
      border: '1px solid var(--panel-edge)', overflow: 'hidden',
    }}>
      <div style={{
        display: 'flex', alignItems: 'center', gap: '6px',
        padding: '6px 10px', background: 'var(--panel)',
        borderBottom: '1px solid var(--panel-edge)',
      }}>
        <Volume2 size={12} style={{ color: 'var(--accent)' }} />
        <span style={{ fontFamily: 'var(--font-mono)', fontWeight: 600, fontSize: '12px', color: 'var(--text-dim)' }}>
          Text to Speech
        </span>
        {voice && <span style={{ fontSize: '11px', color: 'var(--text-faint)' }}>{voice}</span>}
        {format && <span style={{ fontSize: '11px', color: 'var(--text-faint)' }}>.{format}</span>}
      </div>

      <div style={{ padding: 'var(--space-3)' }}>
        {text && (
          <div style={{
            fontSize: '12px', color: 'var(--text-dim)', marginBottom: 'var(--space-2)',
            fontStyle: 'italic', maxHeight: '60px', overflow: 'hidden',
          }}>
            "{text.slice(0, 150)}{text.length > 150 ? '...' : ''}"
          </div>
        )}

        {url ? (
          <>
            <audio ref={audioRef} src={url} preload="metadata" />
            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
              <button
                onClick={togglePlay}
                style={{
                  width: 32, height: 32, borderRadius: '50%',
                  background: 'var(--accent)', border: 'none', cursor: 'pointer',
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  color: 'white', flexShrink: 0,
                }}
              >
                {playing ? <Pause size={14} /> : <Play size={14} style={{ marginLeft: 2 }} />}
              </button>
              <div style={{ flex: 1 }}>
                <div style={{
                  height: 4, background: 'var(--panel-edge)', borderRadius: 2,
                  position: 'relative', overflow: 'hidden',
                }}>
                  <div style={{
                    position: 'absolute', left: 0, top: 0, height: '100%',
                    width: duration ? `${(currentTime / duration) * 100}%` : '0%',
                    background: 'var(--accent)', borderRadius: 2,
                    transition: 'width 0.1s linear',
                  }} />
                </div>
              </div>
              <span style={{ fontFamily: 'var(--font-mono)', fontSize: '11px', color: 'var(--text-faint)', flexShrink: 0 }}>
                {duration ? `${formatTime(currentTime)} / ${formatTime(duration)}` : '--:--'}
              </span>
            </div>
          </>
        ) : (
          <div style={{ textAlign: 'center', color: 'var(--text-faint)', fontSize: '12px' }}>
            {path ? `Saved to: ${path}` : 'Audio generation queued'}
          </div>
        )}
      </div>
    </div>
  );
}
