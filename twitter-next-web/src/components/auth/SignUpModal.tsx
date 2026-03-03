'use client';

import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { X } from 'lucide-react';
import { useState } from 'react';

import { useUIStore } from '@/store/useUIStore';

export function SignUpModal() {
  const isOpen = useUIStore((s) => s.isSignUpModalOpen);
  const onClose = useUIStore((s) => s.closeSignUpModal);
  const openSignInModal = useUIStore((s) => s.openSignInModal);

  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [step, setStep] = useState(1);

    // Reset state when closing
  const handleClose = () => {
    setStep(1);
    setName('');
    setEmail('');
    onClose();
  };

  const handleNext = () => {
    alert('สมัครไม่ได้หรอกโว้ย บังคับ Google login เท่านั้น 5555');
  }

  const handleSignInClick = () => {
    onClose();
    openSignInModal();
  };

  const months = [
    'January', 'February', 'March', 'April', 'May', 'June',
    'July', 'August', 'September', 'October', 'November', 'December'
  ];
  const days = Array.from({ length: 31 }, (_, i) => i + 1);
  const years = Array.from({ length: 120 }, (_, i) => 2024 - i);

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
       <DialogContent showCloseButton={false} className="sm:max-w-[600px] bg-background text-foreground border-border p-0 gap-0 overflow-hidden h-[650px] flex flex-col">
        {/* Header */}
        <div className="flex items-center h-[53px] px-4 shrink-0 relative">
          <button 
            onClick={handleClose}
            className="p-2 rounded-full hover:bg-card transition-colors -ml-2"
          >
             <X size={20} />
          </button>
           {step > 1 && <span className="font-bold text-xl ml-6">Step {step} of 5</span>}
        </div>

        {/* Content */}
        <div className="flex-1 flex flex-col px-20 pt-4 pb-8 overflow-y-auto">
             <div className="w-full max-w-[440px] mx-auto flex flex-col h-full">
            <DialogHeader className="mb-6">
              <DialogTitle className="text-3xl font-bold text-left">
                Create your account
              </DialogTitle>
              <DialogDescription className="sr-only">
                Fill out this form to create a new account.
              </DialogDescription>
            </DialogHeader>

            <div className="flex flex-col gap-6 flex-1">
              <Input
                type="text"
                placeholder="Name"
                 className="bg-background border-border focus-visible:ring-1 focus-visible:ring-primary h-[56px] text-lg"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
              <Input
                type="email"
                placeholder="Email"
                 className="bg-background border-border focus-visible:ring-1 focus-visible:ring-primary h-[56px] text-lg"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
              
              <div className="mt-4">
                <h3 className="font-bold text-[15px] mb-1">Date of birth</h3>
                <p className="text text-muted-foreground text-[14px] mb-4">
                  This will not be shown publicly. Confirm your own age, even if this account is for a business, a pet, or something else.
                </p>
                
                <div className="grid grid-cols-4 gap-3">
                   <div className="col-span-2">
                        <select className="w-full h-[56px] bg-background border border-border rounded-[4px] px-2 text-foreground focus:border-primary outline-none appearance-none">
                            <option value="" disabled selected>Month</option>
                             {months.map(m => <option key={m} value={m}>{m}</option>)}
                        </select>
                   </div>
                    <div className="col-span-1">
                         <select className="w-full h-[56px] bg-background border border-border rounded-[4px] px-2 text-foreground focus:border-primary outline-none appearance-none">
                            <option value="" disabled selected>Day</option>
                             {days.map(d => <option key={d} value={d}>{d}</option>)}
                        </select>
                   </div>
                    <div className="col-span-1">
                        <select className="w-full h-[56px] bg-background border border-border rounded-[4px] px-2 text-foreground focus:border-primary outline-none appearance-none">
                            <option value="" disabled selected>Year</option>
                            {years.map(y => <option key={y} value={y}>{y}</option>)}
                        </select>
                   </div>
                </div>
              </div>
            </div>

            <Button 
                className="cursor-pointer w-full rounded-full bg-foreground text-background hover:bg-foreground/90 h-[52px] font-bold text-[17px] mt-8"
                disabled={!name || !email}
                onClick={handleNext}
            >
              Next
            </Button>
            
            <p className="text-muted-foreground text-[15px] mt-4 text-center">
              Have an account already? <span className="text-primary cursor-pointer hover:underline" onClick={handleSignInClick}>Log in</span>
            </p>

          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
