'use client';

import { useState, useRef, useEffect } from 'react';
import { useAuth } from '@/hooks/useAuth';
import { useUpdateProfile } from '@/hooks/useProfile';
import { Dialog, DialogContent, DialogTitle, DialogDescription } from '@/components/ui/dialog';
import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import { X, Camera } from 'lucide-react';
import type { UserResponse } from '@/types';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { updateProfileSchema, type UpdateProfileInput } from '@/lib/validation';

interface EditProfileModalProps {
  user: UserResponse;
  isOpen: boolean;
  onClose: () => void;
}

export function EditProfileModal({ user, isOpen, onClose }: EditProfileModalProps) {
  const { user: currentUser } = useAuth(); // Need to update global auth store if current user changes

  const {
    register,
    handleSubmit,
    watch,
    reset,
    formState: { errors, isDirty, isValid }
  } = useForm<UpdateProfileInput>({
    resolver: zodResolver(updateProfileSchema),
    defaultValues: {
      displayName: user.displayName || '',
      bio: user.bio || '',
    },
    mode: 'onChange',
  });

  const displayName = watch('displayName') || '';
  const bio = watch('bio') || '';

  const [avatar, setAvatar] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(user.avatarUrl ?? null);
  
  const fileInputRef = useRef<HTMLInputElement>(null);
  const updateMutation = useUpdateProfile();

  const hasChanges = isDirty || avatar !== null;

  useEffect(() => {
    if (isOpen) {
        reset({
            displayName: user.displayName || '',
            bio: user.bio || '',
        });
        setPreviewUrl(user.avatarUrl ?? null);
        setAvatar(null);
    }
  }, [isOpen, user, reset]);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setAvatar(file);
      const url = URL.createObjectURL(file);
      setPreviewUrl(url);
    }
  };

  const onSubmit = async (data: UpdateProfileInput) => {
    try {
      const updatedUser = await updateMutation.mutateAsync({
        displayName: data.displayName || '',
        bio: data.bio || '',
        avatar: avatar ?? undefined,
      });
      
      // Update global auth store if we're editing our own profile
      if (currentUser && currentUser.id === updatedUser.id) {
          // This is a bit of a hack since setAuth requires token, but we just want to update user. 
          // Assuming we can re-fetch or the store has a partial update method (which it likely doesn't).
          // Ideally we invalidate 'auth' query and let it re-fetch.
          // But for now, we rely on React Query invalidation in the hook.
      }
      
      onClose();
    } catch (error) {
      console.error('Failed to update profile:', error);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent showCloseButton={false} className="sm:max-w-[600px] bg-background border-border p-0 gap-0 top-[5%] translate-y-0 sm:top-[10%] min-h-[400px] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between h-[53px] px-4 shrink-0 border-b border-border">
            <div className="flex items-center gap-4">
                <button
                    onClick={onClose}
                    className="p-2 rounded-full hover:bg-card transition-colors -ml-2"
                >
                    <X size={20} className="text-foreground" />
                </button>
                <DialogTitle className="text-[20px] font-bold text-foreground">Edit Profile</DialogTitle>
                <DialogDescription className="sr-only">Make changes to your profile here. Click save when you&apos;re done.</DialogDescription>
            </div>
            <Button
                onClick={handleSubmit(onSubmit)}
                disabled={!displayName.trim() || updateMutation.isPending || !hasChanges || !isValid}
                className="rounded-full bg-foreground text-background hover:bg-foreground/90 h-[32px] font-bold text-[14px] px-4"
            >
                {updateMutation.isPending ? 'Saving...' : 'Save'}
            </Button>
        </div>

        {/* Content */}
        <div className="p-4 flex flex-col gap-6">
            {/* Banner (Placeholder) */}
            <div className="h-[200px] bg-secondary -mt-4 -mx-4 mb-10 relative">
                {/* Avatar Overlay */}
                <div className="absolute -bottom-[40px] left-4">
                     <div className="w-[112px] h-[112px] rounded-full border-4 border-background bg-card relative overflow-hidden group">
                        <Avatar className="w-full h-full rounded-none">
                            <AvatarImage src={previewUrl ?? undefined} alt="" className="object-cover" />
                            <AvatarFallback className="text-[32px] font-bold">{(user.displayName || user.username)[0]}</AvatarFallback>
                        </Avatar>
                        <div className="absolute inset-0 bg-black/30 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer" onClick={() => fileInputRef.current?.click()}>
                            <Camera className="w-6 h-6 text-white" />
                        </div>
                     </div>
                     <input
                        type="file"
                        ref={fileInputRef}
                        className="hidden"
                        accept="image/*"
                        onChange={handleFileChange}
                     />
                </div>
            </div>

            <div className="flex flex-col gap-6 mt-4">
                <div className="border border-border rounded-md px-3 py-1.5 focus-within:border-primary focus-within:ring-1 focus-within:ring-primary transition-colors relative">
                    <div className="flex justify-between items-center mb-0.5">
                        <label className="block text-muted-foreground text-[13px]">Name</label>
                        <span className="text-muted-foreground text-[13px]">{displayName.length} / 30</span>
                    </div>
                    <input 
                        {...register('displayName')}
                        className="w-full bg-transparent text-foreground text-[17px] outline-none"
                    />
                </div>
                {errors.displayName && <span className="text-red-500 text-sm">{errors.displayName.message}</span>}

                <div className="border border-border rounded-md px-3 py-1.5 focus-within:border-primary focus-within:ring-1 focus-within:ring-primary transition-colors relative">
                    <div className="flex justify-between items-center mb-0.5">
                        <label className="block text-muted-foreground text-[13px]">Bio</label>
                        <span className="text-muted-foreground text-[13px]">{bio.length} / 160</span>
                    </div>
                    <textarea 
                        {...register('bio')}
                        className="w-full bg-transparent text-foreground text-[17px] outline-none resize-none h-[80px]"
                    />
                </div>
                {errors.bio && <span className="text-red-500 text-sm">{errors.bio.message}</span>}
            </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
