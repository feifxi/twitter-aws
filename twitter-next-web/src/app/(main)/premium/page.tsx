'use client';

import { useState } from 'react';
import { XLogo } from '@/components/XLogo';
import { Button } from '@/components/ui/button';
import { Check } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";

export default function PremiumPage() {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const features = [
    'เป็นที่รักของแก๊ง call center',
    'โพสต์ไปก็ไม่มีใครอ่าน',
    'เพื่อนไม่รัก',
    'อย่ากดถ้าไม่อยากโดนย่ำกระหม่อม',
  ];

  const handleSubscribe = () => {
    setIsModalOpen(true);
  };

  return (
    <div className="min-h-screen bg-background text-foreground p-4 overflow-y-auto no-scrollbar pb-20">
      <div className="max-w-2xl mx-auto flex flex-col items-center text-center mt-10">
        <h1 className="text-4xl md:text-5xl font-extrabold mb-4">Who are you?</h1>
        <div className="relative mb-8">
            <XLogo className="w-16 h-16 fill-foreground" />
            <div className="absolute -top-2 -right-2 bg-primary rounded-full p-1">
                <Check className="w-4 h-4 text-white" strokeWidth={4} />
            </div>
        </div>
        
        <h2 className="text-2xl font-bold mb-2">Premium</h2>
        <p className="text-muted-foreground mb-8">Detailed verification. Exclusive features.</p>
        
        <div className="grid grid-cols-1 gap-4 w-full mb-8 max-w-md">
            {/* Life Time Plan */}
             <div className="border border-primary rounded-2xl p-6 relative bg-background flex flex-col items-center cursor-pointer hover:bg-card transition-colors">
                 <div className="absolute -top-3 bg-primary text-foreground text-xs font-bold px-2 py-1 rounded">SAVE 80%</div>
                 <h3 className="font-bold text-xl mb-1">Life Time</h3>
                 <p className="text-2xl font-bold mb-4">$99<span className="text-sm font-normal text-muted-foreground">/life time</span></p>
                 <Button 
                    className="cursor-pointer w-full rounded-full font-bold bg-foreground text-background hover:bg-foreground/90" 
                    onClick={handleSubscribe}
                 >
                    Subscribe
                 </Button>
             </div>
        </div>

        <div className="w-full text-left">
            <h3 className="font-bold text-xl mb-4">Everything in Premium</h3>
            <ul className="space-y-4">
                {features.map((feature) => (
                    <li key={feature} className="flex items-center gap-3">
                        <div className="bg-primary rounded-full p-1">
                             <Check className="w-3 h-3 text-white" strokeWidth={4} />
                        </div>
                        <span className="font-medium text-lg">{feature}</span>
                    </li>
                ))}
            </ul>
        </div>
        
        <div className="mt-12 text-muted-foreground text-sm">
            <p className="mb-2">Learn more about Premium and Verified Organizations</p>
        </div>
      </div>

      <Dialog open={isModalOpen} onOpenChange={setIsModalOpen}>
        <DialogContent className="sm:max-w-[400px] border-border bg-background">
          <DialogHeader>
            <DialogTitle className="text-xl font-bold">สมัคร Premium?</DialogTitle>
            <DialogDescription className="text-muted-foreground text-[15px] pt-2">
              ผมรู้ IP คุณแล้ว เตรียมตัวโดนเช็คอินที่หน้าบ้านได้เลย
            </DialogDescription>
          </DialogHeader>
          <DialogFooter className="pt-4">
            <Button 
              className="w-full rounded-full font-bold bg-primary text-white hover:bg-primary/90" 
              onClick={() => setIsModalOpen(false)}
            >
              เข้าใจแล้ว (เตรียมตัว)
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
