'use client';

import { motion } from 'framer-motion';
import { Twitter, Linkedin, Github, Globe, ArrowRight } from 'lucide-react';
import { Link } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';

interface AuthorBioProps {
  name: string;
  bio?: string;
  avatar?: string;
  social?: {
    twitter?: string;
    linkedin?: string;
    github?: string;
    website?: string;
  };
  postsUrl?: string;
}

export function AuthorBio({
  name,
  bio,
  avatar,
  social,
  postsUrl,
}: AuthorBioProps) {
  // Generate initials from name
  const getInitials = (name: string): string => {
    const names = name.trim().split(' ');
    if (names.length >= 2) {
      return `${names[0][0]}${names[names.length - 1][0]}`.toUpperCase();
    }
    return name.charAt(0).toUpperCase();
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4 }}
    >
      <Card className="overflow-hidden rounded-2xl border border-border/50 bg-card/80">
        <CardContent className="p-6">
          <div className="flex flex-col sm:flex-row gap-6">
            {/* Avatar */}
            <div className="flex-shrink-0">
              {avatar ? (
                <img
                  src={avatar}
                  alt={name}
                  className="w-20 h-20 rounded-full object-cover ring-4 ring-border/30"
                />
              ) : (
                <div className="w-20 h-20 rounded-full bg-brand-500/15 flex items-center justify-center text-brand-600 dark:text-brand-400 font-bold text-2xl ring-4 ring-border/30">
                  {getInitials(name)}
                </div>
              )}
            </div>

            {/* Content */}
            <div className="flex-1 min-w-0">
              <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
                <div>
                  <h3 className="text-lg font-semibold mb-1">{name}</h3>
                  {bio && (
                    <p className="text-muted-foreground text-sm leading-relaxed">
                      {bio}
                    </p>
                  )}
                </div>

                {/* Social Links */}
                {social && Object.values(social).some(Boolean) && (
                  <div className="flex items-center gap-2">
                    {social.twitter && (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-9 w-9 rounded-full hover:bg-[#1DA1F2]/10 hover:text-[#1DA1F2]"
                        asChild
                      >
                        <a
                          href={`https://twitter.com/${social.twitter.replace('@', '')}`}
                          target="_blank"
                          rel="noopener noreferrer"
                          aria-label={`${name} on Twitter`}
                        >
                          <Twitter className="h-4 w-4" />
                        </a>
                      </Button>
                    )}
                    {social.linkedin && (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-9 w-9 rounded-full hover:bg-[#0A66C2]/10 hover:text-[#0A66C2]"
                        asChild
                      >
                        <a
                          href={social.linkedin}
                          target="_blank"
                          rel="noopener noreferrer"
                          aria-label={`${name} on LinkedIn`}
                        >
                          <Linkedin className="h-4 w-4" />
                        </a>
                      </Button>
                    )}
                    {social.github && (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-9 w-9 rounded-full hover:bg-muted"
                        asChild
                      >
                        <a
                          href={`https://github.com/${social.github}`}
                          target="_blank"
                          rel="noopener noreferrer"
                          aria-label={`${name} on GitHub`}
                        >
                          <Github className="h-4 w-4" />
                        </a>
                      </Button>
                    )}
                    {social.website && (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-9 w-9 rounded-full hover:bg-muted"
                        asChild
                      >
                        <a
                          href={social.website}
                          target="_blank"
                          rel="noopener noreferrer"
                          aria-label={`${name}'s website`}
                        >
                          <Globe className="h-4 w-4" />
                        </a>
                      </Button>
                    )}
                  </div>
                )}
              </div>

              {/* View all posts link */}
              {postsUrl && (
                <div className="mt-4">
                  <Button
                    variant="outline"
                    size="sm"
                    className="rounded-full gap-1.5"
                    asChild
                  >
                    <Link to={postsUrl}>
                      View all posts
                      <ArrowRight className="h-3.5 w-3.5" />
                    </Link>
                  </Button>
                </div>
              )}
            </div>
          </div>
        </CardContent>
      </Card>
    </motion.div>
  );
}
