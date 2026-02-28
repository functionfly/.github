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

    // Load feedback history
    const loadFeedbackHistory = async () => {
      try {
        const response = await fetch('/api/feedback/history');
        if (response.ok) {
          const history = await response.json();
          setFeedbackHistory(history);
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
        answer: 'We offer multiple support channels including documentation, community forums, email support, and priority support for paid customers.'
      }
    ],
    deployment: [
      {
        question: 'How fast is deployment?',
        answer: 'Deployments typically take 10-30 seconds for cold starts and under 5 seconds for warm deployments.'
      },
      {
        question: 'Can I deploy to multiple regions?',
        answer: 'Yes! FunctionFly allows you to deploy to multiple cloud providers and regions simultaneously for optimal performance and availability.'
      }
    ],
    'getting-started': [
      {
        question: 'How do I get started with FunctionFly?',
        answer: 'Getting started is simple! Sign up for a free account, connect your first cloud provider, and deploy your first function using our CLI or web dashboard.'
      },
      {
        question: 'What programming languages are supported?',
        answer: 'FunctionFly supports all major programming languages including JavaScript/TypeScript, Python, Go, Rust, Java, PHP, Ruby, and .NET.'
      }
    ]
  };

  const handleSubmit = async (e: React.FormEvent) => {
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

      // Add attachments
      attachments.forEach((file, index) => {
        formData.append(`attachment_${index}`, file);
      });

      const response = await fetch('/v1/feedback', {
        method: 'POST',
        body: formData // Remove Content-Type header, let browser set it with boundary
      });

      if (!response.ok) {
        if (response.status === 429) {
          toast.error('Too many submissions. Please try again later.');
        } else {
          toast.error('Failed to submit feedback. Please try again.');
        }
        return;
      }

      // Success handling
      await response.json(); // Parse response to ensure it's valid JSON
      toast.success('Feedback submitted successfully! Thank you for your input.');
      setIsSubmitted(true);

      // Update rate limiting state
      localStorage.setItem('last-feedback-submit', Date.now().toString());

      // Reset form
      setFeedbackType('');
      setSubject('');
      setMessage('');
      setPriority('');
      setBrowserInfo('');
      setAttachments([]);
      setValidationErrors([]);

      // Clear draft from localStorage
      localStorage.removeItem('feedback-draft');

    } catch (error) {
      toast.error('Network error. Please check your connection and try again.');
    } finally {
      setIsSubmitting(false);
    }
  };


  return (
    <div className="min-h-screen bg-bg-primary">
      {/* Navbar */}
      <Navbar variant="landing" />

      {/* Header */}
      <div className="border-b border-border-subtle pt-16">
        <div className="container mx-auto px-4 py-8">
          <div className="flex items-center gap-3 mb-4">
            <MessageSquare className="h-8 w-8 text-text-primary" />
            <h1 className="text-3xl font-bold text-text-primary">Feedback & Support</h1>
          </div>
          <p className="text-text-secondary text-lg">
            Help us improve FunctionFly by sharing your thoughts, reporting bugs, or requesting features.
          </p>
        </div>
      </div>

      <div className="container mx-auto px-4 py-8">
        <div className="max-w-4xl mx-auto space-y-8">

          {/* Draft Restoration */}
          {(feedbackType || subject || message || priority) && (
            <Card className="border-info/20 bg-info/5">
              <CardContent className="pt-6">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <FileText className="h-5 w-5 text-info" />
                    <div>
                      <h3 className="font-semibold text-info">Draft Saved</h3>
                      <p className="text-sm text-text-secondary">
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
            <Card className="border-success/20 bg-success/5">
              <CardContent className="pt-6">
                <div className="flex items-center gap-3">
                  <CheckCircle className="h-6 w-6 text-success" />
                  <div>
                    <h3 className="font-semibold text-success">Feedback Submitted!</h3>
                    <p className="text-sm text-text-secondary">
                      Thank you for your feedback. We'll review it and get back to you if needed.
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Feedback Form */}
          <Card>
            <CardHeader>
              <CardTitle>Submit Feedback</CardTitle>
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
                          className={`p-4 border rounded-lg text-left transition-all hover:border-border-focus ${
                            feedbackType === type.value
                              ? 'border-border-focus bg-brand-500/5 ring-1 ring-border-focus'
                              : 'border-border-subtle hover:bg-bg-hover'
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
                    <SelectTrigger>
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
                  <div className="border-2 border-dashed border-border-default rounded-lg p-4 text-center">
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
                          <div key={index} className="flex items-center gap-2 p-2 bg-bg-tertiary rounded-lg">
                            <File className="h-4 w-4 text-text-muted" />
                            <span className="text-sm flex-1 truncate text-text-primary">{file.name}</span>
                            <span className="text-xs text-text-muted">
                              {(file.size / 1024 / 1024).toFixed(2)}MB
                            </span>
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
                    <div className="h-4 bg-bg-tertiary rounded w-3/4"></div>
                    <div className="h-4 bg-bg-tertiary rounded w-1/2"></div>
                    <div className="h-20 bg-bg-tertiary rounded"></div>
                    <div className="h-12 bg-bg-tertiary rounded"></div>
                  </div>
                )}

                {/* Submit Button - Hidden on mobile, shown on desktop */}
                {!isSubmitting && (
                  <Button
                    type="submit"
                    disabled={isSubmitting}
                    className="w-full hidden md:flex"
                    size="lg"
                  >
                    <Send className="h-4 w-4 mr-2" />
                    Submit Feedback
                  </Button>
                )}

                {/* Sticky submit button on mobile */}
                <div className="fixed bottom-0 left-0 right-0 bg-bg-primary border-t border-border-subtle p-4 md:hidden">
                  <Button
                    type="submit"
                    disabled={isSubmitting}
                    className="w-full"
                    size="lg"
                  >
                    <Send className="h-4 w-4 mr-2" />
                    Submit Feedback
                  </Button>
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
                    <Button variant="outline" size="sm">
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