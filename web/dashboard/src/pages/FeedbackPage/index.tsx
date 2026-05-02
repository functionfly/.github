'use client';

import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { MessageSquare, Send, CheckCircle, Star, FileText, AlertCircle, Lightbulb, Heart, HelpCircle, ChevronDown, Upload, X, Clock, File } from 'lucide-react';
import { toast } from 'sonner';
import { Footer } from '@/pages/LandingPage/components/Footer';
import { Navbar } from '@/components/common/Navbar';

export function FeedbackPage() {
  const textPrimary = 'var(--text-primary)';
  const textSecondary = 'var(--text-secondary)';
  const textMuted = 'var(--text-muted)';
  const bgPrimary = 'var(--bg-primary)';
  const bgSecondary = 'var(--bg-secondary)';
  const bgTertiary = 'var(--bg-tertiary)';
  const bgHover = 'var(--bg-hover)';
  const borderSubtle = 'var(--border-subtle)';
  const borderDefault = 'var(--border-default)';
  const cardBg = 'var(--card)';
  const inputBg = 'var(--input, #ffffff)';
  const inputText = 'var(--input-foreground, #0a0a0a)';
  const brand500 = '#FF6B35';
  const success = 'var(--color-success, #10b981)';
  const warning = 'var(--color-warning, #f59e0b)';
  const error = 'var(--color-error, #ef4444)';
  const info = 'var(--color-info, #3b82f6)';
  const [feedbackType, setFeedbackType] = useState('');
  const [subject, setSubject] = useState('');
  const [message, setMessage] = useState('');
  const [priority, setPriority] = useState('');
  const [browserInfo, setBrowserInfo] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSubmitted, setIsSubmitted] = useState(false);
  const [expandedFaq, setExpandedFaq] = useState<number | null>(null);
  const [attachments, setAttachments] = useState<File[]>([]);
  const [feedbackHistory, setFeedbackHistory] = useState<any[]>([]);
  const [validationErrors, setValidationErrors] = useState<string[]>([]);
  const [lastSubmittedAt, setLastSubmittedAt] = useState<string | null>(null);
  const [uploadProgress, setUploadProgress] = useState<{ [key: string]: number }>({});

  const checkRateLimit = () => {
    const lastSubmit = localStorage.getItem('last-feedback-submit');
    if (lastSubmit && Date.now() - parseInt(lastSubmit) < 3600000) { // 1 hour
      toast.error('Please wait an hour before submitting another feedback');
      return false;
    }
    return true;
  };

  const validateForm = () => {
    const errors: string[] = [];

    if (feedbackType === 'bug' && !message.toLowerCase().includes('steps to reproduce')) {
      errors.push('Bug reports should include steps to reproduce');
    }

    if (feedbackType === 'feature' && message.length < 50) {
      errors.push('Feature requests need more detail (minimum 50 characters)');
    }

    if (message.length > 1000) {
      errors.push('Message must be 1000 characters or less');
    }

    return errors;
  };

  // Auto-save drafts
  useEffect(() => {
    const draft = { feedbackType, subject, message, priority };
    localStorage.setItem('feedback-draft', JSON.stringify(draft));
  }, [feedbackType, subject, message, priority]);

  // Load draft on component mount
  useEffect(() => {
    const draft = localStorage.getItem('feedback-draft');
    if (draft) {
      try {
        const { feedbackType: type, subject: subj, message: msg, priority: pri } = JSON.parse(draft);
        setFeedbackType(type || '');
        setSubject(subj || '');
        setMessage(msg || '');
        setPriority(pri || '');
      } catch (error) {
        // Invalid draft data, ignore
      }
    }

    // Load last submitted timestamp
    const lastSubmit = localStorage.getItem('last-feedback-submit');
    if (lastSubmit) {
      setLastSubmittedAt(new Date(parseInt(lastSubmit)).toLocaleString());
    }

    // Load feedback history
    const loadFeedbackHistory = async () => {
      try {
        const response = await fetch('/api/feedback/history');
        if (response.ok) {
          const data = await response.json();
          setFeedbackHistory(data.feedback || data || []);
        }
      } catch (error) {
        // History loading failed, continue without it
      }
    };

    loadFeedbackHistory();
  }, []);

  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = event.target.files;
    if (!files) return;

    const maxFiles = 5;
    const maxSize = 10 * 1024 * 1024; // 10MB
    const allowedTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp', 'text/plain', 'text/log'];

    const newFiles: File[] = [];
    const errors: string[] = [];

    for (let i = 0; i < files.length; i++) {
      const file = files[i];

      if (attachments.length + newFiles.length >= maxFiles) {
        errors.push(`Maximum ${maxFiles} files allowed`);
        break;
      }

      if (file.size > maxSize) {
        errors.push(`${file.name} is too large (max 10MB)`);
        continue;
      }

      if (!allowedTypes.some(type => file.type === type || file.name.toLowerCase().endsWith(type.split('/')[1]))) {
        errors.push(`${file.name} has unsupported file type`);
        continue;
      }

      newFiles.push(file);
    }

    if (errors.length > 0) {
      toast.error(errors.join('. '));
    }

    if (newFiles.length > 0) {
      setAttachments(prev => [...prev, ...newFiles]);
      toast.success(`Added ${newFiles.length} file(s)`);
    }
  };

  const removeAttachment = (index: number) => {
    setAttachments(prev => prev.filter((_, i) => i !== index));
  };

  const feedbackTypes = [
    { value: 'bug', label: 'Bug Report', icon: AlertCircle, color: 'destructive' },
    { value: 'feature', label: 'Feature Request', icon: Lightbulb, color: 'default' },
    { value: 'improvement', label: 'Improvement', icon: Star, color: 'secondary' },
    { value: 'general', label: 'General Feedback', icon: MessageSquare, color: 'outline' },
  ];

  const priorities = [
    { value: 'low', label: 'Low Priority', description: 'Nice to have' },
    { value: 'medium', label: 'Medium Priority', description: 'Should have' },
    { value: 'high', label: 'High Priority', description: 'Must have' },
    { value: 'critical', label: 'Critical', description: 'Breaking functionality' },
  ];

  // FAQ Integration - Map feedback types to relevant FAQ categories
  const faqMapping: Record<string, string> = {
    bug: 'support',
    feature: 'deployment',
    improvement: 'deployment',
    general: 'getting-started'
  };

  const faqData: Record<string, Array<{ question: string; answer: string }>> = {
    support: [
      {
        question: 'How do I report a bug?',
        answer: 'When reporting a bug, please include steps to reproduce, expected vs. actual behavior, and your browser/OS information. This helps us identify and fix issues quickly.'
      },
      {
        question: 'What kind of support do you offer?',
        answer: 'We offer documentation, GitHub community discussions, email support at api-support@functionfly.com, and priority support for paid customers.'
      }
    ],
    deployment: [
      {
        question: 'How fast is deployment?',
        answer: 'FunctionFly deploys to 35+ edge locations worldwide. Cold starts vary by runtime: Go functions are fastest, while Python/JavaScript may take longer depending on dependencies. Warm deployments are typically faster.'
      },
      {
        question: 'Can I deploy to multiple regions?',
        answer: 'Yes! FunctionFly deploys to 35+ edge locations across North America, Europe, Asia, and Australia for optimal performance and availability worldwide.'
      }
    ],
    'getting-started': [
      {
        question: 'How do I get started with FunctionFly?',
        answer: 'Sign up for a free account, install the ffly CLI (go install github.com/functionfly/functionfly/cmd/ffly@latest), connect your first cloud provider, and deploy your first function using ffly deploy.'
      },
      {
        question: 'What programming languages are supported?',
        answer: 'FunctionFly supports JavaScript/TypeScript, Python, and Go. Additional runtimes available include Rust (WASM), Swift (WASM), Kotlin (WASM), C/C++ (WASM), and Ruby (mruby).'
      }
    ]
  };

  const handleSubmit = async (e: React.FormEvent, retryCount = 0) => {
    e.preventDefault();

    if (!feedbackType || !subject || !message) {
      toast.error('Please fill in all required fields');
      return;
    }

    // Smart form validation
    const errors = validateForm();
    if (errors.length > 0) {
      setValidationErrors(errors);
      toast.error('Please fix the validation errors below');
      return;
    }

    // Clear any previous validation errors
    setValidationErrors([]);

    // Check rate limit before proceeding
    if (!checkRateLimit()) {
      return;
    }

    setIsSubmitting(true);

    try {
      const formData = new FormData();
      formData.append('feedbackType', feedbackType);
      formData.append('subject', subject);
      formData.append('message', message);
      formData.append('priority', priority);
      formData.append('browserInfo', browserInfo);

      // Track upload progress for attachments
      const uploadProgressState: { [key: string]: number } = {};
      attachments.forEach((file, index) => {
        formData.append(`attachment_${index}`, file);
        uploadProgressState[`attachment_${index}`] = 0;
      });
      setUploadProgress(uploadProgressState);

      const response = await fetch('/v1/feedback', {
        method: 'POST',
        body: formData
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        const errorMessage = errorData.error || 'Failed to submit feedback';

        if (response.status === 429) {
          toast.error('Too many submissions. Please try again later.');
        } else if (response.status >= 500 && retryCount < 2) {
          // Retry on server errors (up to 2 retries)
          toast.warning(`Server error, retrying... (${retryCount + 1}/2)`);
          await new Promise(resolve => setTimeout(resolve, 1000 * (retryCount + 1)));
          setIsSubmitting(false);
          return handleSubmit(e, retryCount + 1);
        } else {
          toast.error(errorMessage);
        }
        return;
      }

      // Success handling
      const result = await response.json();
      toast.success('Feedback submitted successfully! Thank you for your input.');
      setIsSubmitted(true);

      // Update rate limiting state and last submitted timestamp
      const now = Date.now();
      localStorage.setItem('last-feedback-submit', now.toString());
      setLastSubmittedAt(new Date(now).toLocaleString());

      // Reset form
      setFeedbackType('');
      setSubject('');
      setMessage('');
      setPriority('');
      setBrowserInfo('');
      setAttachments([]);
      setValidationErrors([]);
      setUploadProgress({});

      // Clear draft from localStorage
      localStorage.removeItem('feedback-draft');

      // Show view status link with feedback ID if available
      if (result?.id) {
        toast.success(<div className="flex items-center gap-2">
          <span>Your feedback ID: {result.id}</span>
        </div>, { duration: 10000 });
      }

    } catch (error) {
      if (retryCount < 2) {
        toast.warning('Network error, retrying...');
        await new Promise(resolve => setTimeout(resolve, 1000 * (retryCount + 1)));
        setIsSubmitting(false);
        return handleSubmit(e, retryCount + 1);
      }
      toast.error('Network error. Please check your connection and try again.');
    } finally {
      setIsSubmitting(false);
    }
  };


  return (
    <div className="min-h-screen" style={{ backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}>
      {/* Navbar */}
      <Navbar variant="landing" />

      {/* Header */}
      <div className="border-b border-border-subtle pt-16" style={{ backgroundColor: 'var(--bg-primary)' }}>
        <div className="container mx-auto px-4 py-8">
          <div className="flex items-center gap-3 mb-4">
            <MessageSquare className="h-8 w-8" style={{ color: 'var(--text-primary)' }} />
            <h1 className="text-3xl font-bold" style={{ color: 'var(--text-primary)' }}>Feedback & Support</h1>
          </div>
          <p style={{ color: 'var(--text-secondary)' }}>
            Help us improve FunctionFly by sharing your thoughts, reporting bugs, or requesting features.
          </p>
        </div>
      </div>

      <div className="container mx-auto px-4 py-8">
        <div className="max-w-4xl mx-auto space-y-8" style={{ color: textPrimary }}>

          {/* Draft Restoration */}
          {(feedbackType || subject || message || priority) && (
            <Card style={{ borderColor: 'var(--color-info)', backgroundColor: 'rgba(59, 130, 246, 0.05)' }}>
              <CardContent className="pt-6">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <FileText className="h-5 w-5" style={{ color: 'var(--color-info)' }} />
                    <div>
                      <h3 className="font-semibold" style={{ color: 'var(--color-info)' }}>Draft Saved</h3>
                      <p className="text-sm" style={{ color: textSecondary }}>
                        Your progress is automatically saved as you type
                      </p>
                    </div>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setFeedbackType('');
                      setSubject('');
                      setMessage('');
                      setPriority('');
                      setBrowserInfo('');
                      setAttachments([]);
                      setValidationErrors([]);
                      localStorage.removeItem('feedback-draft');
                      toast.success('Draft cleared');
                    }}
                  >
                    Clear Draft
                  </Button>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Success Message */}
          {isSubmitted && (
            <Card style={{ borderColor: 'var(--color-success)', backgroundColor: 'rgba(16, 185, 129, 0.05)' }}>
              <CardContent className="pt-6">
                <div className="flex items-center gap-3">
                  <CheckCircle className="h-6 w-6" style={{ color: 'var(--color-success)' }} />
                  <div className="flex-1">
                    <h3 className="font-semibold" style={{ color: 'var(--color-success)' }}>Feedback Submitted!</h3>
                    <p className="text-sm" style={{ color: textSecondary }}>
                      Thank you for your feedback. We'll review it and get back to you if needed.
                    </p>
                    {lastSubmittedAt && (
                      <p className="text-xs mt-1" style={{ color: 'var(--text-muted)' }}>
                        Submitted at {lastSubmittedAt}
                      </p>
                    )}
                  </div>
                </div>
                {lastSubmittedAt && (
                  <div className="mt-4 pt-4 border-t" style={{ borderColor: 'var(--color-success)' }}>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        setIsSubmitted(false);
                      }}
                    >
                      Submit Another
                    </Button>
                  </div>
                )}
              </CardContent>
            </Card>
          )}

          {/* Feedback Form */}
          <Card style={{ backgroundColor: cardBg, borderColor: borderDefault }}>
            <CardHeader>
              <CardTitle style={{ color: textPrimary }}>Submit Feedback</CardTitle>
            </CardHeader>
            <CardContent className="pb-24 md:pb-6">
              <form onSubmit={handleSubmit} className="space-y-6">
                {/* Feedback Type */}
                <div className="space-y-3">
                  <label className="text-sm font-medium">
                    Feedback Type <span className="text-red-500">*</span>
                  </label>
                  <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
                    {feedbackTypes.map((type) => {
                      const IconComponent = type.icon;
                      return (
                        <button
                          key={type.value}
                          type="button"
                          onClick={() => {
                            setFeedbackType(type.value);
                            setExpandedFaq(null); // Reset FAQ expansion when changing feedback type
                          }}
                          className={`p-4 border rounded-lg text-left transition-all hover:border-[var(--border-focus)] ${
                            feedbackType === type.value
                              ? 'border-[var(--border-focus)] bg-brand-500/5 ring-1 ring-[var(--border-focus)]'
                              : 'border-[var(--border-subtle)] hover:bg-[var(--bg-hover)]'
                          }`}
                        >
                          <IconComponent className="h-5 w-5 mb-2 text-text-primary" />
                          <div className="font-medium text-sm">{type.label}</div>
                        </button>
                      );
                    })}
                  </div>
                </div>

                {/* Subject */}
                <div className="space-y-2">
                  <label htmlFor="subject" className="text-sm font-medium">
                    Subject <span className="text-red-500">*</span>
                  </label>
                  <Input
                    id="subject"
                    value={subject}
                    onChange={(e) => setSubject(e.target.value)}
                    placeholder="Brief description of your feedback"
                    required
                  />
                </div>

                {/* Priority */}
                <div className="space-y-2">
                  <label className="text-sm font-medium">Priority Level</label>
                  <Select value={priority} onValueChange={setPriority}>
                    <SelectTrigger
                      style={{
                        backgroundColor: 'var(--input, #ffffff)',
                        borderColor: 'var(--border-default, rgba(0,0,0,0.18))',
                        color: 'var(--input-foreground, #0a0a0a)',
                      }}
                    >
                      <SelectValue placeholder="Select priority (optional)" />
                    </SelectTrigger>
                    <SelectContent>
                      {priorities.map((p) => (
                        <SelectItem key={p.value} value={p.value}>
                          <div className="flex items-center gap-2">
                            <span>{p.label}</span>
                            <span className="text-xs text-muted-foreground">({p.description})</span>
                          </div>
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                {/* Browser/OS Info - Progressive Form Enhancement */}
                {feedbackType === 'bug' && (
                  <div className="space-y-2">
                    <label htmlFor="browserInfo" className="text-sm font-medium">
                      Browser/OS Info
                    </label>
                    <Input
                      id="browserInfo"
                      value={browserInfo}
                      onChange={(e) => setBrowserInfo(e.target.value)}
                      placeholder="Chrome 120.0.6099.109 on macOS Sonoma"
                    />
                    <p className="text-xs text-muted-foreground">
                      Include your browser version and operating system to help us reproduce the issue.
                    </p>
                  </div>
                )}

                {/* Message */}
                <div className="space-y-2">
                  <div className="flex justify-between items-center">
                    <label htmlFor="message" className="text-sm font-medium">
                      Message <span className="text-red-500">*</span>
                    </label>
                    <span className={`text-xs ${message.length > 900 ? 'text-red-500' : 'text-muted-foreground'}`}>
                      {message.length}/1000
                    </span>
                  </div>
                  <Textarea
                    id="message"
                    value={message}
                    onChange={(e) => setMessage(e.target.value.slice(0, 1000))}
                    placeholder="Please provide detailed information about your feedback, including steps to reproduce if it's a bug, or specific suggestions for features/improvements."
                    rows={6}
                    required
                  />
                  <p className="text-xs text-muted-foreground">
                    Be as specific as possible to help us understand and address your feedback effectively.
                  </p>
                </div>

                {/* Attachments */}
                <div className="space-y-2">
                  <label className="text-sm font-medium">Attachments (optional)</label>
                  <div className="border-2 border-dashed border-[var(--border-default)] rounded-lg p-4 text-center">
                    <input
                      type="file"
                      multiple
                      accept="image/*,.txt,.log"
                      className="hidden"
                      id="file-upload"
                      onChange={handleFileUpload}
                    />
                    <label htmlFor="file-upload" className="cursor-pointer">
                      <Upload className="h-8 w-8 mx-auto mb-2 text-text-muted" />
                      <p className="text-sm text-text-muted">Click to upload screenshots or files</p>
                      <p className="text-xs text-text-muted mt-1">Max 5 files, 10MB each (images, .txt, .log)</p>
                    </label>
                  </div>

                  {/* Display attached files */}
                  {attachments.length > 0 && (
                    <div className="space-y-2">
                      <p className="text-sm font-medium">Attached files:</p>
                      <div className="space-y-2">
                        {attachments.map((file, index) => (
                          <div key={index} className="flex items-center gap-2 p-2 bg-[var(--bg-tertiary)] rounded-lg">
                            <File className="h-4 w-4 text-text-muted" />
                            <span className="text-sm flex-1 truncate text-text-primary">{file.name}</span>
                            <span className="text-xs text-text-muted">
                              {(file.size / 1024 / 1024).toFixed(2)}MB
                            </span>
                            {uploadProgress[`attachment_${index}`] !== undefined && uploadProgress[`attachment_${index}`] < 100 && (
                              <span className="text-xs text-info">
                                {uploadProgress[`attachment_${index}`]}%
                              </span>
                            )}
                            <Button
                              type="button"
                              variant="ghost"
                              size="sm"
                              onClick={() => removeAttachment(index)}
                              className="h-6 w-6 p-0"
                            >
                              <X className="h-3 w-3" />
                            </Button>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>

                {/* Validation Errors */}
                {validationErrors.length > 0 && (
                  <div className="space-y-2">
                    <div className="bg-error/10 border border-error/20 rounded-lg p-3">
                      <h4 className="text-sm font-medium text-error mb-2">
                        Please fix the following issues:
                      </h4>
                      <ul className="text-sm text-error space-y-1">
                        {validationErrors.map((error, index) => (
                          <li key={index}>• {error}</li>
                        ))}
                      </ul>
                    </div>
                  </div>
                )}

                {/* Loading States - Skeleton Loading */}
                {isSubmitting && (
                  <div className="space-y-4 animate-pulse">
                    <div className="h-4 bg-[var(--bg-tertiary)] rounded w-3/4"></div>
                    <div className="h-4 bg-[var(--bg-tertiary)] rounded w-1/2"></div>
                    <div className="h-20 bg-[var(--bg-tertiary)] rounded"></div>
                    <div className="h-12 bg-[var(--bg-tertiary)] rounded"></div>
                  </div>
                )}

                {/* Submit Button - Hidden on mobile, shown on desktop */}
                {!isSubmitting && (
                  <button
                    type="submit"
                    disabled={isSubmitting}
                    className="w-full hidden md:flex items-center justify-center gap-2 h-12 px-6 rounded-lg text-white font-semibold transition-all hover:brightness-110 active:scale-[0.98]"
                    style={{
                      background: 'linear-gradient(135deg, #FF6B35 0%, #FF4F5E 100%)',
                      boxShadow: '0 4px 14px rgba(255, 107, 53, 0.4)',
                      color: '#ffffff',
                    }}
                  >
                    <Send className="h-4 w-4" />
                    Submit Feedback
                  </button>
                )}

                {/* Sticky submit button on mobile */}
                <div className="fixed bottom-0 left-0 right-0 bg-[var(--bg-primary)] border-t border-[var(--border-subtle)] p-4 md:hidden">
                  <button
                    type="submit"
                    disabled={isSubmitting}
                    className="w-full flex items-center justify-center gap-2 h-12 px-6 rounded-lg text-white font-semibold transition-all hover:brightness-110 active:scale-[0.98]"
                    style={{
                      background: 'linear-gradient(135deg, #FF6B35 0%, #FF4F5E 100%)',
                      boxShadow: '0 4px 14px rgba(255, 107, 53, 0.4)',
                      color: '#ffffff',
                    }}
                  >
                    <Send className="h-4 w-4" />
                    Submit Feedback
                  </button>
                </div>
              </form>
            </CardContent>
          </Card>

          {/* Feedback Guidelines */}
          <Card>
            <CardHeader>
              <CardTitle>Feedback Guidelines</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <h4 className="font-semibold mb-2 flex items-center gap-2">
                    <AlertCircle className="h-4 w-4" />
                    Bug Reports
                  </h4>
                  <ul className="text-sm text-muted-foreground space-y-1">
                    <li>• Include steps to reproduce the issue</li>
                    <li>• Describe expected vs. actual behavior</li>
                    <li>• Mention your browser/OS if relevant</li>
                    <li>• Attach screenshots if possible</li>
                  </ul>
                </div>
                <div>
                  <h4 className="font-semibold mb-2 flex items-center gap-2">
                    <Lightbulb className="h-4 w-4" />
                    Feature Requests
                  </h4>
                  <ul className="text-sm text-muted-foreground space-y-1">
                    <li>• Explain the problem this solves</li>
                    <li>• Describe how you would use it</li>
                    <li>• Mention similar features you've seen</li>
                    <li>• Consider implementation complexity</li>
                  </ul>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* FAQ Integration */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <HelpCircle className="h-5 w-5" />
                {feedbackType ? 'Related Questions' : 'Frequently Asked Questions'}
              </CardTitle>
              <p className="text-sm text-muted-foreground">
                {feedbackType
                  ? `Here are some frequently asked questions that might help with your ${feedbackTypes.find(t => t.value === feedbackType)?.label.toLowerCase()}.`
                  : 'Select a feedback type above to see related questions, or browse our general FAQ.'
                }
              </p>
            </CardHeader>
            <CardContent className="space-y-3">
              {(feedbackType && faqMapping[feedbackType] && faqData[faqMapping[feedbackType]])
                ? faqData[faqMapping[feedbackType]].map((faq: { question: string; answer: string }, index: number) => (
                    <div key={index} className="border border-border-subtle rounded-lg overflow-hidden">
                      <button
                        onClick={() => setExpandedFaq(expandedFaq === index ? null : index)}
                        className="w-full px-4 py-3 text-left flex items-center justify-between hover:bg-bg-hover transition-colors"
                      >
                        <div className="flex items-center gap-3">
                          <HelpCircle className="h-4 w-4 text-text-primary shrink-0" />
                          <span className="font-medium text-sm text-text-primary">{faq.question}</span>
                        </div>
                        <ChevronDown
                          className={`h-4 w-4 text-text-muted transition-transform ${
                            expandedFaq === index ? 'rotate-180' : ''
                          }`}
                        />
                      </button>
                      {expandedFaq === index && (
                        <div className="px-4 pb-3 border-t border-border-subtle">
                          <p className="text-sm text-text-secondary pt-3 leading-relaxed">
                            {faq.answer}
                          </p>
                        </div>
                      )}
                    </div>
                  ))
                : faqData['getting-started'].map((faq: { question: string; answer: string }, index: number) => (
                    <div key={index} className="border border-border-subtle rounded-lg overflow-hidden">
                      <button
                        onClick={() => setExpandedFaq(expandedFaq === index ? null : index)}
                        className="w-full px-4 py-3 text-left flex items-center justify-between hover:bg-bg-hover transition-colors"
                      >
                        <div className="flex items-center gap-3">
                          <HelpCircle className="h-4 w-4 text-text-primary shrink-0" />
                          <span className="font-medium text-sm text-text-primary">{faq.question}</span>
                        </div>
                        <ChevronDown
                          className={`h-4 w-4 text-text-muted transition-transform ${
                            expandedFaq === index ? 'rotate-180' : ''
                          }`}
                        />
                      </button>
                      {expandedFaq === index && (
                        <div className="px-4 pb-3 border-t border-border-subtle">
                          <p className="text-sm text-text-secondary pt-3 leading-relaxed">
                            {faq.answer}
                          </p>
                        </div>
                      )}
                    </div>
                  ))
              }
            </CardContent>
          </Card>

          {/* Feedback History */}
          {feedbackHistory.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Clock className="h-5 w-5" />
                  Your Previous Feedback
                </CardTitle>
                <p className="text-sm text-muted-foreground">
                  Track the status of your previous submissions
                </p>
              </CardHeader>
              <CardContent className="space-y-3">
                {feedbackHistory.slice(0, 5).map((item: any, index: number) => (
                  <div key={index} className="border border-border rounded-lg p-4">
                    <div className="flex items-start justify-between mb-2">
                      <div className="flex items-center gap-2">
                        <div className={`w-2 h-2 rounded-full ${
                          item.status === 'resolved' ? 'bg-success' :
                          item.status === 'in-progress' ? 'bg-warning' :
                          'bg-text-muted'
                        }`} />
                        <span className="font-medium text-sm capitalize text-text-primary">{item.feedbackType}</span>
                      </div>
                      <span className="text-xs text-muted-foreground">
                        {new Date(item.createdAt).toLocaleDateString()}
                      </span>
                    </div>
                    <h4 className="font-medium text-sm mb-1">{item.subject}</h4>
                    <p className="text-sm text-muted-foreground line-clamp-2">{item.message}</p>
                    <div className="flex items-center gap-2 mt-2">
                      <span className={`text-xs px-2 py-1 rounded-full ${
                        item.status === 'resolved' ? 'bg-success/10 text-success' :
                        item.status === 'in-progress' ? 'bg-warning/10 text-warning' :
                        'bg-text-muted/10 text-text-muted'
                      }`}>
                        {item.status || 'submitted'}
                      </span>
                      {item.priority && (
                        <span className="text-xs text-muted-foreground">
                          Priority: {item.priority}
                        </span>
                      )}
                    </div>
                  </div>
                ))}
                {feedbackHistory.length > 5 && (
                  <p className="text-sm text-muted-foreground text-center">
                    And {feedbackHistory.length - 5} more previous submissions...
                  </p>
                )}
              </CardContent>
            </Card>
          )}

          {/* Alternative Contact */}
          <Card>
            <CardHeader>
              <CardTitle>Other Ways to Reach Us</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="flex items-start gap-3">
                  <MessageSquare className="h-5 w-5 text-text-primary mt-0.5" />
                  <div>
                    <h4 className="font-semibold text-text-primary">Community Forum</h4>
                    <p className="text-sm text-text-secondary mb-2">
                      Join discussions with other users and the FunctionFly team.
                    </p>
                    <Button variant="outline" size="sm">
                      Visit Forum
                    </Button>
                  </div>
                </div>
                <div className="flex items-start gap-3">
                  <Heart className="h-5 w-5 text-text-primary mt-0.5" />
                  <div>
                    <h4 className="font-semibold text-text-primary">Documentation</h4>
                    <p className="text-sm text-text-secondary mb-2">
                      Check our docs for answers to common questions.
                    </p>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => window.open(import.meta.env.DEV ? '/docs' : 'https://docs.functionfly.com', '_blank')}
                    >
                      View Docs
                    </Button>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Back to Home */}
          <div className="text-center">
            <Link to="/">
              <Button variant="outline">
                <FileText className="h-4 w-4 mr-2" />
                Back to Home
              </Button>
            </Link>
          </div>
        </div>
      </div>

      {/* Footer */}
      <Footer />
    </div>
  );
}