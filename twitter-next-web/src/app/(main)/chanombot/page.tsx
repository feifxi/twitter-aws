'use client';

import { useState, useRef, useEffect } from 'react';
import { Image as ImageIcon, Send, Sparkles } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { useAuth } from '@/hooks/useAuth';
import { useUIStore } from '@/store/useUIStore';

type Message = {
  id: string;
  role: 'user' | 'assistant';
  content: string;
};

const INITIAL_MESSAGES: Message[] = [
  {
    id: '1',
    role: 'assistant',
    content: "Greetings, human! I am ChanomBot, engineered with a healthy dose of sarcasm and a direct line to the universe's most mildly interesting facts. What's on your mind today?",
  },
];

export default function ChanomBotPage() {
  const { user, isLoggedIn, accessToken } = useAuth();
  const { openSignInModal } = useUIStore();
  const [messages, setMessages] = useState<Message[]>(INITIAL_MESSAGES);
  const [input, setInput] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSend = async () => {
    if (!input.trim()) return;

    if (!isLoggedIn) {
      openSignInModal();
      return;
    }
    
    const userContent = input.trim();
    const newUserMsg: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: userContent,
    };
    
    setMessages((prev) => [...prev, newUserMsg]);
    setInput('');

    // Prepare history for Gemini format
    const history = messages.map(m => ({
      role: m.role === 'assistant' ? 'model' : 'user',
      text: m.content
    }));

    // Placeholder for bot message
    const botMsgId = (Date.now() + 1).toString();
    const initialBotMsg: Message = {
      id: botMsgId,
      role: 'assistant',
      content: '',
    };
    setMessages((prev) => [...prev, initialBotMsg]);

    try {
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/assistant`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${accessToken}`,
        },
        body: JSON.stringify({
          query: userContent,
          history: history,
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to fetch chat response');
      }

      const reader = response.body?.getReader();
      if (!reader) throw new Error('No reader available');

      const decoder = new TextDecoder();
      let done = false;
      let accumulatedContent = '';
      let sseBuffer = '';

      while (!done) {
        const { value, done: doneReading } = await reader.read();
        done = doneReading;
        sseBuffer += decoder.decode(value, { stream: !done });

        // Parse SSE frames: each frame is "data: <content>\n\n"
        const frames = sseBuffer.split('\n\n');
        // Keep the last incomplete frame in the buffer
        sseBuffer = frames.pop() || '';

        for (const frame of frames) {
          for (const line of frame.split('\n')) {
            if (line.startsWith('event: error')) {
              // Skip error event marker, data line follows
              continue;
            }
            if (line.startsWith('data: ')) {
              accumulatedContent += line.slice(6);
            }
          }
        }

        setMessages((prev) => 
          prev.map((msg) => 
            msg.id === botMsgId ? { ...msg, content: accumulatedContent } : msg
          )
        );
      }
    } catch (error) {
      console.error('Chat error:', error);
      setMessages((prev) => 
        prev.map((msg) => 
          msg.id === botMsgId ? { ...msg, content: 'Error: Failed to get response from ChanomBot.' } : msg
        )
      );
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <div className="flex flex-col h-screen max-w-full relative">
      {/* Header */}
      <header className="sticky top-0 z-10 flex items-center justify-between px-4 py-3 bg-background/80 backdrop-blur-md border-b border-border">
        <div className="flex items-center gap-4">
          <div className="w-8 h-8 rounded-full bg-primary flex items-center justify-center">
            <Sparkles className="w-5 h-5 text-white" />
          </div>
          <div>
            <h1 className="text-xl font-bold leading-tight flex items-center gap-2">
              ChanomBot
              <span className="text-xs font-normal px-2 py-0.5 bg-accent text-accent-foreground rounded-full">Beta</span>
            </h1>
            <p className="text-xs text-muted-foreground leading-tight">Fun Mode</p>
          </div>
        </div>
      </header>

      {/* Messages Area */}
      <div className="flex-1 overflow-y-auto px-4 py-6 space-y-6 no-scrollbar pb-32">
        {messages.map((msg) => (
          <div key={msg.id} className="flex gap-4">
            {msg.role === 'assistant' ? (
              <div className="w-10 h-10 shrink-0 rounded-full bg-primary flex items-center justify-center">
                <Sparkles className="w-6 h-6 text-white" />
              </div>
            ) : (
              <Avatar className="w-10 h-10 shrink-0 border border-border/50">
                <AvatarImage src={user?.avatarUrl ?? undefined} />
                <AvatarFallback>{user?.displayName?.[0] || 'U'}</AvatarFallback>
              </Avatar>
            )}
            <div className="flex flex-col flex-1 min-w-0">
              <div className="flex items-center gap-2 mb-1">
                <span className="font-bold text-[15px]">
                  {msg.role === 'assistant' ? 'ChanomBot' : user?.displayName || 'You'}
                </span>
                {msg.role === 'user' && (
                  <span className="text-muted-foreground text-[15px]">
                    @{user?.username || 'user'}
                  </span>
                )}
              </div>
              <div className="text-[15px] leading-relaxed whitespace-pre-wrap">
                {msg.content}
              </div>
            </div>
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>

      {/* Input Area */}
      <div className="absolute bottom-0 left-0 right-0 bg-background/90 backdrop-blur-md pt-2 pb-4 px-4 border-t border-border">
        <div className="max-w-4xl mx-auto rounded-2xl bg-accent/30 border border-border/50 focus-within:border-primary/50 focus-within:ring-1 focus-within:ring-primary/50 transition-all">
          <textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Ask ChanomBot anything..."
            className="w-full max-h-48 min-h-[60px] bg-transparent resize-none outline-none p-4 text-[15px] placeholder:text-muted-foreground"
            rows={1}
            style={{ height: 'auto' }}
            onInput={(e) => {
              const target = e.target as HTMLTextAreaElement;
              target.style.height = 'auto';
              target.style.height = `${Math.min(target.scrollHeight, 200)}px`;
            }}
          />
          <div className="flex items-center justify-between p-2 pt-0">
            <div className="flex items-center gap-1">
              <Button size="icon" variant="ghost" className="rounded-full text-primary hover:bg-primary/10 h-9 w-9">
                <ImageIcon className="w-5 h-5" />
              </Button>
            </div>
            <Button 
              size="icon" 
              onClick={handleSend}
              disabled={!input.trim()}
              className="rounded-full bg-primary text-primary-foreground hover:bg-primary/90 h-9 w-9 disabled:opacity-50"
            >
              <Send className="w-4 h-4" />
            </Button>
          </div>
        </div>
        <div className="text-center mt-3">
          <p className="text-xs text-muted-foreground">ChanomBot can make mistakes. Consider verifying important information.</p>
        </div>
      </div>
    </div>
  );
}
