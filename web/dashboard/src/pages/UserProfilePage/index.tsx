import { useParams, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { User, Calendar, Package, ExternalLink, AlertCircle } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Navbar } from "@/components/common/Navbar";
import { usersApi } from "@/api/users";
import type { PublicUserProfile } from "@/types";

function ProfileSkeleton() {
  return (
    <div className="animate-pulse space-y-6">
      <div className="flex items-center gap-6">
        <div className="w-24 h-24 rounded-full bg-bg-secondary" />
        <div className="space-y-3 flex-1">
          <div className="h-6 bg-bg-secondary rounded w-48" />
          <div className="h-4 bg-bg-secondary rounded w-32" />
          <div className="h-4 bg-bg-secondary rounded w-64" />
        </div>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="h-32 bg-bg-secondary rounded-lg" />
        ))}
      </div>
    </div>
  );
}

function FunctionCard({ fn }: { fn: PublicUserProfile["publishedFunctions"][0] }) {
  return (
    <Card className="bg-bg-secondary border-border-subtle hover:border-brand-500/50 transition-colors">
      <CardHeader className="pb-2">
        <div className="flex items-start justify-between gap-2">
          <CardTitle className="text-sm font-medium text-text-primary truncate">
            {fn.name}
          </CardTitle>
          <Link
            to={`/fx/${fn.author}/${fn.name}`}
            className="shrink-0 text-text-muted hover:text-brand-400 transition-colors"
          >
            <ExternalLink className="w-4 h-4" />
          </Link>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        <p className="text-xs text-text-secondary line-clamp-2">{fn.description}</p>
        <div className="flex items-center justify-between">
          <Badge variant="secondary" className="text-xs">
            v{fn.version}
          </Badge>
          {fn.executionCount !== undefined && (
            <span className="text-xs text-text-muted">
              {fn.executionCount.toLocaleString()} runs
            </span>
          )}
        </div>
        {fn.tags && fn.tags.length > 0 && (
          <div className="flex flex-wrap gap-1">
            {fn.tags.slice(0, 3).map((tag) => (
              <span
                key={tag}
                className="text-xs px-1.5 py-0.5 rounded bg-brand-500/10 text-brand-400"
              >
                {tag}
              </span>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export function UserProfilePage() {
  const { username } = useParams<{ username: string }>();

  const {
    data: profile,
    isLoading,
    isError,
    error,
  } = useQuery<PublicUserProfile>({
    queryKey: ["user-profile", username],
    queryFn: () => usersApi.getPublicProfile(username!),
    enabled: !!username,
    staleTime: 5 * 60 * 1000,
    retry: 1,
  });

  const joinedDate = profile?.createdAt
    ? new Date(profile.createdAt).toLocaleDateString("en-US", {
        year: "numeric",
        month: "long",
      })
    : null;

  return (
    <div className="min-h-screen bg-bg-primary">
      <Navbar variant="landing" />

      <main className="max-w-4xl mx-auto px-4 pt-24 pb-16">
        {isLoading && <ProfileSkeleton />}

        {isError && (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="flex flex-col items-center justify-center py-24 text-center"
          >
            <AlertCircle className="w-12 h-12 text-text-muted mb-4" />
            <h1 className="text-2xl font-bold text-text-primary mb-2">
              User not found
            </h1>
            <p className="text-text-secondary mb-6">
              {(error as Error)?.message?.includes("404")
                ? `No user with username "@${username}" exists.`
                : "Failed to load this profile. Please try again."}
            </p>
            <Link to="/registry">
              <Button variant="outline">Browse Functions</Button>
            </Link>
          </motion.div>
        )}

        {profile && (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4 }}
            className="space-y-8"
          >
            {/* Profile Header */}
            <div className="flex flex-col sm:flex-row items-start sm:items-center gap-6">
              {/* Avatar */}
              <div className="w-24 h-24 rounded-full bg-linear-to-br from-brand-500 to-brand-600 flex items-center justify-center text-white text-3xl font-bold shrink-0 overflow-hidden">
                {profile.avatar ? (
                  <img
                    src={profile.avatar}
                    alt={profile.name || profile.username}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  (profile.name || profile.username).charAt(0).toUpperCase()
                )}
              </div>

              {/* Info */}
              <div className="flex-1 min-w-0">
                <h1 className="text-2xl font-bold text-text-primary">
                  {profile.name || profile.username}
                </h1>
                <p className="text-brand-400 font-medium">@{profile.username}</p>
                {profile.bio && (
                  <p className="text-text-secondary mt-2 text-sm">{profile.bio}</p>
                )}
                <div className="flex items-center gap-4 mt-3 text-xs text-text-muted">
                  {joinedDate && (
                    <span className="flex items-center gap-1">
                      <Calendar className="w-3.5 h-3.5" />
                      Joined {joinedDate}
                    </span>
                  )}
                  <span className="flex items-center gap-1">
                    <Package className="w-3.5 h-3.5" />
                    {profile.publishedFunctions.length} function
                    {profile.publishedFunctions.length !== 1 ? "s" : ""}
                  </span>
                </div>
              </div>
            </div>

            {/* Published Functions */}
            <section>
              <h2 className="text-lg font-semibold text-text-primary mb-4 flex items-center gap-2">
                <Package className="w-5 h-5 text-brand-500" />
                Published Functions
              </h2>

              {profile.publishedFunctions.length === 0 ? (
                <div className="text-center py-12 text-text-muted">
                  <User className="w-10 h-10 mx-auto mb-3 opacity-40" />
                  <p>No published functions yet.</p>
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {profile.publishedFunctions.map((fn) => (
                    <FunctionCard key={`${fn.author}/${fn.name}`} fn={fn} />
                  ))}
                </div>
              )}
            </section>
          </motion.div>
        )}
      </main>
    </div>
  );
}
