import { useContext } from 'react';
import { SupportChatContext } from './SupportChat';

export function useSupportChat() {
  const context = useContext(SupportChatContext);
  if (!context) {
    throw new Error('useSupportChat must be used within SupportChatProvider');
  }
  return context;
}
