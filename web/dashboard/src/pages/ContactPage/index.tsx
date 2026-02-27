'use client';

import { Link } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Navbar } from '@/components/common/Navbar';
import { Footer } from '@/pages/LandingPage/components';
import { MessageSquare, Send } from 'lucide-react';

// Validation and types
import { contactFormSchema, ContactFormData } from './validation';

// Hooks
import { useContactFormAutoSave, useContactFormSubmission } from './hooks';

// Components
import {
  AutoSaveIndicator,
  TextFormField,
  EmailFormField,
  SelectFormField,
  TextareaFormField,
  SubmitStatusAlerts,
  FormValidationSummary,
  ContactMethods,
  OfficeInformation
} from './components';

export function ContactPage() {
  // React Hook Form setup
  const form = useForm<ContactFormData>({
    resolver: zodResolver(contactFormSchema),
    mode: 'onChange', // Enable real-time validation
    defaultValues: {
      name: '',
      email: '',
      subject: '',
      message: '',
      category: 'general'
    }
  });

  const {
    register,
    handleSubmit,
    formState: { errors, isValid, touchedFields, dirtyFields },
    watch
  } = form;

  // Custom hooks
  const { lastSaved, showDraftIndicator, clearDraft } = useContactFormAutoSave(form);
  const { isSubmitting, submitStatus, onSubmit } = useContactFormSubmission(form);

  // Helper function to get field state
  const getFieldState = (fieldName: keyof ContactFormData) => {
    const hasError = !!errors[fieldName];
    const isTouched = !!touchedFields[fieldName];
    const isDirty = !!dirtyFields[fieldName];
    const hasValue = !!watch(fieldName);

    return {
      hasError,
      isTouched,
      isDirty,
      hasValue,
      isValid: !hasError && isTouched && hasValue,
      showError: hasError && isTouched
    };
  };

  const categories = [
    { value: 'general', label: 'General Inquiry' },
    { value: 'technical', label: 'Technical Support' },
    { value: 'billing', label: 'Billing & Pricing' },
    { value: 'sales', label: 'Sales & Enterprise' },
    { value: 'partnership', label: 'Partnerships' },
    { value: 'feedback', label: 'Feedback & Suggestions' },
  ];

  return (
    <div className="contact-page">
      <Navbar variant="landing" />

      <div className="container mx-auto px-4 py-8 pt-20 relative z-10">
        <div className="max-w-6xl mx-auto grid grid-cols-1 lg:grid-cols-2 gap-8">

          {/* Contact Form */}
          <div className="space-y-6">
            <div className="contact-form-container animate-float">
              <div className="contact-form-header">
                <div className="contact-form-icon">
                  <Send className="h-6 w-6" />
                </div>
                <h2 className="contact-form-title">
                  Send us a Message
                </h2>
                <p className="contact-form-description">
                  We'd love to hear from you. Fill out the form below and we'll get back to you soon.
                </p>
              </div>

              <AutoSaveIndicator
                lastSaved={lastSaved}
                showDraftIndicator={showDraftIndicator}
                onClearDraft={clearDraft}
              />

              <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
                <div className="contact-form-grid">
                  <TextFormField
                    id="name"
                    label="Full Name"
                    required
                    placeholder="Your full name"
                    fieldState={getFieldState('name')}
                    error={errors.name}
                    register={register('name')}
                  />

                  <EmailFormField
                    id="email"
                    label="Email Address"
                    required
                    placeholder="your.email@example.com"
                    fieldState={getFieldState('email')}
                    error={errors.email}
                    register={register('email')}
                  />
                </div>

                <SelectFormField
                  id="category"
                  label="Category"
                  fieldState={getFieldState('category')}
                  error={errors.category}
                  register={register('category')}
                  options={categories}
                />

                <TextFormField
                  id="subject"
                  label="Subject"
                  required
                  placeholder="Brief description of your inquiry"
                  fieldState={getFieldState('subject')}
                  error={errors.subject}
                  register={register('subject')}
                />

                <TextareaFormField
                  id="message"
                  label="Message"
                  required
                  placeholder="Please provide details about your inquiry..."
                  fieldState={getFieldState('message')}
                  error={errors.message}
                  register={register('message')}
                  characterCount={{
                    current: watch('message')?.length || 0,
                    max: 2000
                  }}
                />

                <SubmitStatusAlerts submitStatus={submitStatus} />

                <button
                  type="submit"
                  disabled={isSubmitting || !isValid}
                  className="contact-submit-button"
                >
                  <span className="flex items-center justify-center gap-2">
                    <Send className="h-4 w-4" />
                    Send Message
                  </span>
                </button>

                <FormValidationSummary
                  errors={errors}
                  touchedFields={touchedFields}
                  isValid={isValid}
                />
              </form>
            </div>
          </div>

          {/* Contact Information */}
          <div className="space-y-6">
            <ContactMethods />

            <OfficeInformation />

            {/* FAQ Link */}
            <div className="contact-info-card animate-float">
              <div className="text-center">
                <div className="contact-info-card-icon">
                  <MessageSquare className="h-5 w-5" />
                </div>
                <h4 className="font-semibold mb-3 text-white">Looking for quick answers?</h4>
                <p className="text-sm text-text-secondary mb-6 leading-relaxed">
                  Check out our frequently asked questions for instant answers to common questions.
                </p>
                <Link to="/faq" className="inline-block">
                  <button className="faq-contact-button secondary">
                    <MessageSquare className="h-4 w-4 mr-2" />
                    View FAQ
                  </button>
                </Link>
              </div>
            </div>
          </div>
        </div>

      </div>

      <Footer />
    </div>
  );
}
