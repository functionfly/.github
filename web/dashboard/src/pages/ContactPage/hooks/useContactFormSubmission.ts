import { useState } from 'react';
import { UseFormReturn } from 'react-hook-form';
import { ContactFormData } from '../validation';

export function useContactFormSubmission(form: UseFormReturn<ContactFormData>) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitStatus, setSubmitStatus] = useState<'idle' | 'success' | 'error'>('idle');

  const onSubmit = async (data: ContactFormData) => {
    setIsSubmitting(true);
    setSubmitStatus('idle');

    try {
      // Simulate form submission
      await new Promise(resolve => setTimeout(resolve, 2000));

      setSubmitStatus('success');
      // Note: Draft clearing is handled by the auto-save hook
    } catch (error) {
      setSubmitStatus('error');
    } finally {
      setIsSubmitting(false);
    }
  };

  return {
    isSubmitting,
    submitStatus,
    onSubmit
  };
}